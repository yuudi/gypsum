package cli

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"

	"github.com/yuudi/gypsum/gypsum"
)

type zeroConfig struct {
	NickName      []string
	CommandPrefix string
	SuperUsers    []string
}

type Config struct {
	Host        string
	Port        int
	AccessToken string
	LogLevel    string
	ZeroBot     zeroConfig
	Gypsum      gypsum.ConfigType
}

func (c *Config) Save() error {
	tmpl, err := template.New("config").Parse(defaultConfig)
	if err != nil {
		return err
	}
	file, err := os.Create("gypsum_config.toml")
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) CheckValid() error {
	changed, err := c.Gypsum.CheckValid()
	if err != nil {
		return err
	}
	if changed {
		return c.Save()
	}
	return nil
}

//go:embed default_config.toml.tmpl
var defaultConfig string

func initialConfig(interactive bool) {
	config := Config{
		Host:        "127.0.0.1",
		Port:        6700,
		AccessToken: "",
		LogLevel:    "INFO",
		ZeroBot: zeroConfig{
			NickName:      []string{"机器人", "笨蛋"},
			CommandPrefix: "",
			SuperUsers:    []string{},
		},
		Gypsum: gypsum.ConfigType{
			Listen:         "http://0.0.0.0:9900",
			Password:       "",
			ExternalAssets: "",
			ResourceShare:  "file",
			HttpBackRef:    "",
		},
	}
	if interactive {
		config.Host = promptEnter("正向 ws 连接主机", "如果 onebot 也在此计算机上则为默认值 127.0.0.1", "127.0.0.1", nil)
		config.Port, _ = strconv.Atoi(promptEnter("正向 ws 端口", "onebot 中设置的正向 ws 端口", "6700", isDigit))
		config.AccessToken = promptEnter("正向 ws 连接密钥", "onebot 中设置的正向 ws 连接密钥，如果没有设置则为空", "", nil)
		config.Gypsum.Password = promptEnter("网页控制台密码", "设置用于登录网页控制台的密码", "", func(s string) bool {
			return len(s) > 0
		})
		advancedSetting := promptEnter("开始高级设置？( y/[N] )", "", "N", nil)
		if advancedSetting == "y" {
			//TODO
			fmt.Println("暂无")
		}
	}
	if err := config.Save(); err != nil {
		fmt.Println("无法生成配置文件：错误：", err)
		os.Exit(1)
	}
	fmt.Println("配置文件已生成。")
}

func readConfig() (*Config, error) {
	fileContent, err := os.ReadFile("gypsum_config.toml")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("配置文件不存在，将生成默认配置文件")
			initialConfig(false)
			fmt.Println("请修改配置文件后再启动。")
			fmt.Println("提示：执行 gypsum init -i 可以帮助你生成配置文件")
		} else {
			fmt.Printf("无法读取配置文件：错误%s\n", err)
		}
		return nil, err
	}
	fileContent = bytes.TrimPrefix(fileContent, []byte{0xef, 0xbb, 0xbf}) // remove utf-8 BOM
	var conf Config
	if _, err := toml.Decode(string(fileContent), &conf); err != nil {
		fmt.Printf("无法解析配置文件：错误%s\n", err)
		return nil, err
	}
	return &conf, conf.CheckValid()
}

func promptEnter(name, help, defaults string, validator func(string) bool) string {
	fmt.Printf("输入%s：\n%s\n回车则为默认值（\"%s\"）\n%s>", name, help, defaults, name)
	for {
		in := readline()
		if len(in) == 0 {
			return defaults
		}
		if validator != nil && !validator(in) {
			fmt.Printf("请输入正确的值\n%s>", name)
			continue
		}
		return in
	}
}
func readline() string {
	buf := bufio.NewReader(os.Stdin)
	line, err := buf.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(line)
}

func isDigit(s string) bool {
	_, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return true
}
