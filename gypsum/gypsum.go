package gypsum

import (
	"log"
	"os"

	zero "github.com/wdvxdr1123/ZeroBot"
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

func init() {
	zero.RegisterPlugin(&gypsumPlugin{}) // 注册插件
}

type gypsumPlugin struct{}

func (_ *gypsumPlugin) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "yuudi",
		PluginName: "石膏自定义",
		Version:    "v" + os.Getenv("GYPSUM_VERSION"),
		Details:    "石膏自定义",
	}
}

func (_ *gypsumPlugin) Start() { // 插件主体
	if err := initTemplating(); err != nil {
		log.Printf("lua引擎初始化错误：%s", err)
		return
	}
	if err := initDb(); err != nil {
		log.Printf("数据库初始化错误：%s", err)
		return
	}
	if err := loadData(); err != nil {
		log.Printf("数据库加载错误：%s", err)
		return
	}
	go serveWeb()
}
