package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/yuudi/gypsum/gypsum"
)

var (
	version = "dev"
	commit  = "none"
)

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
		if runtime.GOOS == "windows" {
			fmt.Println("按回车键关闭")
			_, _ = fmt.Scanln()
		}
		return
	}
	lv, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		lv = log.InfoLevel
	}
	log.SetLevel(lv)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		DisableQuote:    true,
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
	})
	gypsum.Config = conf.Gypsum
	zero.Run(zero.Option{
		Host:          conf.Host,
		Port:          strconv.Itoa(conf.Port),
		AccessToken:   conf.AccessToken,
		NickName:      conf.ZeroBot.NickName,
		CommandPrefix: conf.ZeroBot.CommandPrefix,
		SuperUsers:    conf.ZeroBot.SuperUsers,
	})

	select {}
}
