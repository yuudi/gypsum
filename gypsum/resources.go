package gypsum

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Resource struct {
	FileName  string `json:"file_name"`
	Ext       string `json:"ext"`
	Sha256Sum string `json:"sha256_sum"`
}

var resources map[uint64]Resource
var resDir string // absolute path of resource directory

func (r *Resource) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(r); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ResourceFromBytes(b []byte) (*Resource, error) {
	r := &Resource{}
	buffer := bytes.Buffer{}
	buffer.Write(b)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(r)
	return r, err
}
func loadResources() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	resDir = path.Join(pwd, "resources")
	if resourceDir, err := os.Stat(resDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(resDir, 0644); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		if !resourceDir.IsDir() {
			panic("resource directory exists and is not directory")
		}
	}
	resources = make(map[uint64]Resource)
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-resources-")), nil)
	defer func() {
		iter.Release()
		if err = iter.Error(); err != nil {
			log.Errorf("载入数据错误：%s", err)
		}
	}()
	for iter.Next() {
		key := ToUint(iter.Key()[17:])
		value := iter.Value()
		r, e := ResourceFromBytes(value)
		if e != nil {
			log.Errorf("无法加载资源%d：%s", key, e)
			continue
		}
		resources[key] = *r
	}
}

func resourcePath(filename string) string {
	return path.Join(resDir, filename)
}

func resourceIDByHash(sum string) (uint64, bool) {
	if len(sum) != 64 {
		return 0, false
	}
	b, err := hex.DecodeString(sum)
	if err != nil {
		return 0, false
	}
	v, err := db.Get(append([]byte("gypsum-resources_hash-"), b[:]...), nil)
	if err != nil {
		if err != leveldb.ErrNotFound {
			log.Errorf("error reading database: %s", err)
		}
		return 0, false
	}
	return ToUint(v), true
}

func getResources(c *gin.Context) {
	c.JSON(200, resources)
}

func getResourceByID(c *gin.Context) {
	resourceIDStr := c.Param("rid")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 64)
	if err != nil {
		resourceID, ok := resourceIDByHash(resourceIDStr)
		if !ok {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "resource not found",
			})
			return
		} else {
			c.Redirect(302, fmt.Sprintf("/api/v1/resources/%d", resourceID))
			return
		}
	}
	r, ok := resources[resourceID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	c.JSON(200, r)
}

func downloadResource(c *gin.Context) {
	resourceIDStr := c.Param("rid")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	r, ok := resources[resourceID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	// c.Header("Content-Description", "File Transfer")
	// c.Header("Content-Transfer-Encoding", "binary")
	// c.Header("Content-Disposition", "attachment; filename="+r.FileName+r.Ext)
	// c.Header("Content-Type", "application/octet-stream")
	// c.File(path.Join(resDir, r.Sha256Sum+r.Ext))
	c.FileAttachment(path.Join(resDir, r.Sha256Sum+r.Ext), r.FileName+r.Ext)
}

func uploadResource(c *gin.Context) {
	fileFullName := c.Param("name")
	nameSplit := strings.Split(fileFullName, ".")
	var fileName, ext string
	if len(nameSplit) == 1 {
		ext = ""
		fileName = nameSplit[0]
	} else {
		ext = "." + nameSplit[len(nameSplit)-1]
		fileName = fileFullName[:len(fileFullName)-len(ext)]
	}
	bodyReader := c.Request.Body
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    6000,
			"message": fmt.Sprintf("error when reading request body: %s", err),
		})
	}
	hashBytes := sha256.Sum256(body)
	hashHex := hex.EncodeToString(hashBytes[:])
	if err := ioutil.WriteFile(path.Join(resDir, hashHex+ext), body, 0444); err != nil {
		c.JSON(500, gin.H{
			"code":    6000,
			"message": fmt.Sprintf("error when writing file: %s", err),
		})
	}
	cursor++
	if err := db.Put([]byte("gypsum-$meta-cursor"), U64ToBytes(cursor), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	// check if resource already exist
	idx, err := db.Get(append([]byte("gypsum-resources_hash-"), hashBytes[:]...), nil)
	if err == nil {
		// already exist
		c.JSON(200, gin.H{
			"code":        1,
			"message":     "already exist",
			"resource_id": ToUint(idx),
		})
		return
	} else {
		if err != leveldb.ErrNotFound {
			// error other than "ErrNotFound"
			c.JSON(500, gin.H{
				"code":    3000,
				"message": fmt.Sprintf("Server got itself into trouble: %s", err),
			})
			return
		}
	}
	// not exist, go on
	resource := Resource{
		FileName:  fileName,
		Ext:       ext,
		Sha256Sum: hashHex,
	}
	v, err := resource.ToBytes()
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	if err = db.Put(append([]byte("gypsum-resources-"), U64ToBytes(cursor)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	if err = db.Put(append([]byte("gypsum-resources_hash-"), hashBytes[:]...), U64ToBytes(cursor), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	resources[cursor] = resource
	c.JSON(201, gin.H{
		"code":        0,
		"message":     "ok",
		"resource_id": cursor,
	})
}

func deleteResource(c *gin.Context) {
	resourceIDStr := c.Param("rid")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	_, ok := resources[resourceID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	delete(resources, resourceID)
	if err := db.Delete(append([]byte("gypsum-resources-"), U64ToBytes(resourceID)...), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3001,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
}

type NamePatch struct {
	FileName string `json:"file_name"`
}

func renameResource(c *gin.Context) {
	resourceIDStr := c.Param("rid")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	r, ok := resources[resourceID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such resource",
		})
		return
	}
	np := NamePatch{}
	if err := c.BindJSON(&np); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	r.FileName = np.FileName
	v, err := r.ToBytes()
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	if err = db.Put(append([]byte("gypsum-resources-"), U64ToBytes(resourceID)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
}
