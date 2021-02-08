package main

import (
	_ "embed"

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
