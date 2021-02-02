package main

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
	Gypsum struct {
		Listen   string
		Username string
		Password string
	}
}

var defaultConfig = `# gypsum 配置文件
# 正向 ws 服务端主机
Host = "127.0.0.1"

# 正向 ws 服务端端口
Port = 6700

# 正向 ws 令牌码
AccessToken = ""

# 日志级别
LogLevel = "INFO"

[Gypsum]
# 网页控制台监听地址与端口
Listen = "0.0.0.0:9900"

# 控制台账号
Username = "admin"

# 控制台密码
Password = "set-your-password-here"

[ZeroBot]
# BOT 昵称，叫昵称等同于 @BOT
NickName = ["机器人"]

# 命令前缀，建议留空
CommandPrefix = ""

# 主人，gypsum 用不到，可留空
SuperUsers = [""]
`
