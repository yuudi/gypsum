package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/yuudi/gypsum/gypsum"
)

var (
	version = "dev"
	commit  = "none"
)

type Config struct {
	Host        string
	Port        int
	AccessToken string
	Gypsum      struct {
		Listen   string
		Username string
		Password string
	}
}

const defaultConfig = `
Host = "127.0.0.1"  # 正向 ws 服务端主机
Port = 6700  # 正向 ws 服务端端口
AccessToken = ""  # 正向 ws 令牌码

[Gypsum]
Listen = "0.0.0.0:9900"  # 网页控制台监听地址与端口
Username = "admin"  # 控制台账号
Password = "set-your-password-here"  # 控制台密码
`

func main() {
	fmt.Printf("gypsum %s, commit %s\n\n", version, commit)
	var conf Config
	if _, err := toml.DecodeFile("gypsum_config.toml", &conf); err != nil {
		if os.IsNotExist(err) {
			if err := ioutil.WriteFile("gypsum_config.toml", []byte(defaultConfig), 0644); err != nil {
				fmt.Printf("无法生成配置文件：错误%s\n", err)
			} else {
				fmt.Println("配置文件已生成，请修改配置文件后再启动")
			}
		} else {
			fmt.Printf("无法读取配置文件：错误%s\n", err)
		}
		return

	}
	gypsum.Config = conf.Gypsum
	zero.Run(zero.Option{
		Host:          conf.Host,
		Port:          strconv.Itoa(conf.Port),
		AccessToken:   conf.AccessToken,
		NickName:      []string{},
		CommandPrefix: "\000",
		SuperUsers:    []string{},
	})

	select {}
}
