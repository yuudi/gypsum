package gypsum

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"

	_ "github.com/yuudi/gypsum/gypsum/helper/jsoniter_plugin_integer_interface"
)

type ConfigType struct {
	Listen         string
	Username       string
	Password       string
	ExternalAssets string
	ResourceShare  string
	HttpBackRef    string
}

func (c ConfigType) checkValid() error {
	switch c.ResourceShare {
	case "file": // doing nothing
	case "http":
		strings.TrimSuffix(c.HttpBackRef, "/")
	default:
		return errors.New("unknown ResourceShare: " + c.ResourceShare)
	}
	return nil
}

var Config = ConfigType{
	Listen:         "0.0.0.0:9900",
	Username:       "admin",
	Password:       "admin",
	ExternalAssets: "",
	ResourceShare:  "",
	HttpBackRef:    "",
}

var (
	BuildVersion = "0.0.0-unknown"
	BuildCommit  = "unknown"
)

func init() {
	zero.RegisterPlugin(&gypsumPlugin{}) // 注册插件
}

type gypsumPlugin struct{}

func (_ *gypsumPlugin) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "yuudi",
		PluginName: "石膏自定义",
		Version:    "v" + BuildVersion,
		Details:    "石膏自定义",
	}
}

func (_ *gypsumPlugin) Start() { // 插件主体
	if err := initTemplating(); err != nil {
		log.Errorf("pongo2引擎初始化错误：%s", err)
		return
	}
	if err := initDb(); err != nil {
		log.Errorf("数据库初始化错误：%s", err)
		return
	}
	if err := loadData(); err != nil {
		log.Errorf("数据库加载错误：%s", err)
		return
	}
	initWeb()
}

func getGypsumVersion(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": BuildVersion,
		"commit":  BuildCommit,
	})
}
