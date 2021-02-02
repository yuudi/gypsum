package gypsum

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"

	_ "github.com/yuudi/gypsum/helper/jsoniter_plugin_integer_interface"
)

var Config = struct {
	Listen   string
	Username string
	Password string
}{
	"0.0.0.0:9900",
	"admin",
	"admin",
}

var (
	gypsumVersion = "0.0.0-dev"
	gypsumCommit  = "none"
)

func init() {
	fmt.Printf("gypsum %s, commit %s\n\n", gypsumVersion, gypsumCommit)
	zero.RegisterPlugin(&gypsumPlugin{}) // 注册插件
}

type gypsumPlugin struct{}

func (_ *gypsumPlugin) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "yuudi",
		PluginName: "石膏自定义",
		Version:    "v" + gypsumVersion,
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
	go serveWeb()
}
