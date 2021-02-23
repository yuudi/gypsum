package gypsum

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"

	_ "github.com/yuudi/gypsum/gypsum/helper/jsoniter_plugin_integer_interface"
)

type ConfigType struct {
	Listen         string
	Password       string
	PasswordSalt   string
	ExternalAssets string
	ResourceShare  string
	HttpBackRef    string
}

func (c *ConfigType) CheckValid() (changed bool, err error) {
	switch c.ResourceShare {
	case "file": // doing nothing
	case "http":
		if strings.HasSuffix(c.HttpBackRef, "/") {
			c.HttpBackRef = c.HttpBackRef[:len(c.HttpBackRef)-1]
			changed = true
		}
	default:
		return false, errors.New("unknown ResourceShare: " + c.ResourceShare)
	}
	if len(c.Password) == 0 {
		return false, errors.New("未设置密码")
	}
	if len(c.PasswordSalt) == 0 {
		salt := make([]byte, 12)
		if _, err = rand.Read(salt); err != nil {
			return
		}
		c.PasswordSalt = base64.StdEncoding.EncodeToString(salt)
		// although `salt` is bytes, we do not use it, but the string
		passwordEncrypted := sha256.Sum256([]byte(c.Password + c.PasswordSalt))
		c.Password = hex.EncodeToString(passwordEncrypted[:])
		changed = true
	}
	return
}

var Config *ConfigType

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
		PluginName: "冰石自定义",
		Version:    "v" + BuildVersion,
		Details:    "冰石自定义",
	}
}

func (_ *gypsumPlugin) Start() { // 插件主体
	if err := initTemplating(); err != nil {
		log.Fatalf("pongo2引擎初始化错误：%s", err)
		return
	}
	if err := initDb(); err != nil {
		log.Fatalf("数据库初始化错误：%s", err)
		return
	}
	if err := loadData(); err != nil {
		log.Fatalf("数据库加载错误：%s", err)
		return
	}
	initWeb()
}
