package gypsum

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/go-update"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func UpdateGypsum(newVersion, mirror string, forcedUpdate bool, logger func(...interface{})) error {
	// ensure old version path
	if err := os.MkdirAll("old", 0644); err != nil {
		return err
	}
	var apiUrl string
	var getReleaseFirstItem = false
	switch newVersion {
	case "", "latest", "stable":
		apiUrl = "https://api.github.com/repos/yuudi/gypsum/releases/latest"
	case "beta":
		apiUrl = "https://api.github.com/repos/yuudi/gypsum/releases?per_page=1"
		getReleaseFirstItem = true
	default:
		newVersion = strings.TrimPrefix(newVersion, "v")
		apiUrl = "https://api.github.com/repos/yuudi/gypsum/releases/tags/v" + newVersion
	}
	logger("fetching version " + newVersion)
	res, err := http.Get(apiUrl)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return errors.New("cannot find version: " + newVersion)
		}
		return errors.New(fmt.Sprintf("error response status: %d", res.StatusCode))
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	release := gjson.ParseBytes(resBody)
	if getReleaseFirstItem {
		release = release.Array()[0]
	}
	if !forcedUpdate && release.Get("tag_name").String()[1:] == BuildVersion {
		return errors.New("same version, no need to update")
	}
	var goos, goarch, archiveExt, exeExt, binaryDownloadAddress string
	switch runtime.GOOS {
	case "darwin":
		goos = "mac"
	default:
		goos = runtime.GOOS
	}
	switch runtime.GOARCH {
	case "386":
		goarch = "x86"
	case "amd64":
		goarch = "x86_64"
	default:
		goarch = runtime.GOARCH
	}
	if runtime.GOOS == "windows" {
		archiveExt = ".zip"
		exeExt = ".exe"
	} else {
		archiveExt = ".tar.gz"
		exeExt = ""
	}
	assetNameSuffix := goos + "-" + goarch + archiveExt
	logger(fmt.Sprintf("finding asset for %s, %s %s", newVersion, goos, goarch))
	for _, asset := range release.Get("assets").Array() {
		if strings.HasSuffix(asset.Get("name").Str, assetNameSuffix) {
			binaryDownloadAddress = asset.Get("browser_download_url").Str
			break
		}
	}
	if len(binaryDownloadAddress) == 0 {
		return errors.New(fmt.Sprintf("cannot find asset for %s, %s %s", newVersion, goos, goarch))
	}
	if len(mirror) != 0 {
		binaryDownloadAddress = strings.Replace(binaryDownloadAddress, "github.com", mirror, -1)
	}
	logger("downloading from " + binaryDownloadAddress)
	binaryDownload, err := http.Get(binaryDownloadAddress)
	if err != nil {
		return err
	}
	defer binaryDownload.Body.Close()
	if binaryDownload.StatusCode != 200 {
		return errors.New(fmt.Sprintf("error response status: %d", binaryDownload.StatusCode))
	}
	logger("uncompressing")
	var binaryExeReader io.Reader
	if runtime.GOOS == "windows" {
		resBody, err = io.ReadAll(binaryDownload.Body)
		if err != nil {
			return err
		}
		zipReader, err := zip.NewReader(bytes.NewReader(resBody), int64(len(resBody)))
		if err != nil {
			return err
		}
		for _, file := range zipReader.File {
			if file.Name == "gypsum.exe" {
				fileReader, err := file.Open()
				if err != nil {
					return err
				}
				binaryExeReader = fileReader
				break
			}
		}
		if binaryExeReader == nil {
			return errors.New("executable binary not found in zip file")
		}
	} else {
		uncompressedStream, err := gzip.NewReader(binaryDownload.Body)
		if err != nil {
			return err
		}
		tarReader := tar.NewReader(uncompressedStream)
		for {
			header, err := tarReader.Next()
			if err != nil {
				if err == io.EOF {
					return errors.New("executable binary not found in tar.gz file")
				}
				return err
			}
			if header.Name == "gypsum" {
				binaryExeReader = tarReader
				break
			}
		}
	}
	err = update.Apply(binaryExeReader, update.Options{
		OldSavePath: "old/gypsum-" + BuildVersion + "-" + time.Now().Format("20060102-150405") + exeExt,
	})
	if err != nil {
		return err
	}
	logger("succeed!")
	return nil
}

var updateStatus struct {
	Lock           sync.Once
	Updating       bool
	LastRunMessage string
}

type updateRequest struct {
	NewVersion   string `json:"new_version"`
	Mirror       string `json:"mirror"`
	ForcedUpdate bool   `json:"forced_update"`
}

func getUpdateStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":     0,
		"updating": updateStatus.Updating,
		"message":  updateStatus.LastRunMessage,
	})
}

func requestUpdateGypsum(c *gin.Context) {
	running := true
	updateStatus.Lock.Do(func() {
		running = false
	})
	if running {
		c.JSON(409, gin.H{
			"code":    1,
			"message": "an updating job is processing",
		})
		return
	}
	updateStatus.Updating = true
	updateStatus.LastRunMessage = "processing"
	var req updateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		updateStatus.Lock = sync.Once{} // new `sync.Once` so to run again
		updateStatus.Updating = false
		return
	}
	c.JSON(201, gin.H{
		"code":    0,
		"message": "updating started",
	})
	go func() {
		log.Debug(req.NewVersion, req.Mirror)
		err := UpdateGypsum(req.NewVersion, req.Mirror, req.ForcedUpdate, log.Info)
		if err != nil {
			log.Error(err)
			updateStatus.Lock = sync.Once{} // new `sync.Once` so to run again
			updateStatus.Updating = false
			updateStatus.LastRunMessage = err.Error()
			return
		}
		log.Info("updating complete, restarting")
		os.Exit(5) // restart
	}()
}
