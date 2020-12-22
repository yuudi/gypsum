package gypsum

import (
	"log"

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
		Version:    "0.2.0",
		Details:    "石膏自定义",
	}
}

func (_ *gypsumPlugin) Start() { // 插件主体
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
