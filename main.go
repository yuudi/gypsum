package main

import "github.com/yuudi/gypsum/cli"

var (
	version = "0.0.0-unknown"
	commit  = "unknown"
)

func main() {
	cli.Entry(version, commit)
}
