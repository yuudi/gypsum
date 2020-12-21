package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	zero "github.com/wdvxdr1123/ZeroBot"
	gypsum "github.com/yuudi/gypsum/gypsum"
)

type Config struct {
	Host        string
	Port        int
	AccessToken string
	Listen      string
}

const defaultConfig = `
Host = "127.0.0.1"
Port = 6700
AccessToken = ""
Listen = "0.0.0.0:9900"
`

func main() {
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
	gypsum.Listen = conf.Listen
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
