package cli

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/yuudi/gypsum/gypsum"
)

type Config struct {
	Host        string
	Port        int
	AccessToken string
	LogLevel    string
	ZeroBot     struct {
		NickName      []string
		CommandPrefix string
		SuperUsers    []string
	}
	Gypsum gypsum.ConfigType
}

//go:embed default_config.toml
var defaultConfig string

func initialConfig() {
	if err := os.WriteFile("gypsum_config.toml", []byte(defaultConfig), 0644); err != nil {
		fmt.Printf("无法生成配置文件：错误%s\n", err)
	} else {
		fmt.Println("配置文件已生成。")
	}
}

func readConfig() (*Config, error) {
	var conf Config
	fileContent, err := os.ReadFile("gypsum_config.toml")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("配置文件不存在，将生成默认配置文件")
			initialConfig()
			fmt.Println("请修改配置文件后再启动。")
		} else {
			fmt.Printf("无法读取配置文件：错误%s\n", err)
		}
		return nil, err
	}
	fileContent = bytes.TrimPrefix(fileContent, []byte{0xef, 0xbb, 0xbf}) // remove utf-8 BOM
	if _, err := toml.Decode(string(fileContent), &conf); err != nil {
		fmt.Printf("无法解析配置文件：错误%s\n", err)
		return nil, err
	}
	return &conf, nil
}
