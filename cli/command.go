package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	"github.com/yuudi/gypsum/gypsum"
)

var (
	version = "0.0.0-unknown"
	commit  = "unknown"
)

type commandOptions struct {
	action        string
	updateVersion string
	githubMirror  string
	updateForced  bool
	extractPath   string
	interactive   bool
}

func parseCommand() commandOptions {
	var cmd commandOptions
	app := kingpin.New("gypsum", "gypsum cli")
	app.Command("daemon", "start daemon gypsum").Default()
	app.Command("run", "start run gypsum")
	cmdInit := app.Command("init", "initial gypsum configuration file")
	cmdInit.Flag("interactive", "interactive help to initial config file").Default("false").Short('i').BoolVar(&cmd.interactive)
	cmdExtract := app.Command("extract-web", "extract web assets from gypsum")
	cmdExtract.Arg("path", "path to save web assets").Default(".").StringVar(&cmd.extractPath)
	cmdUpdate := app.Command("update", "update gypsum")
	cmdUpdate.Arg("version", "new version to fetch").Default("stable").StringVar(&cmd.updateVersion)
	cmdUpdate.Flag("mirror", "mirror to replace github.com for downloading").Short('m').StringVar(&cmd.githubMirror)
	cmdUpdate.Flag("force", "forced update").Short('f').Default("false").BoolVar(&cmd.updateForced)
	app.Version(fmt.Sprintf("gypsum %s, commit %s", version, commit))
	app.VersionFlag.Short('V')
	app.HelpFlag.Short('h')
	cmd.action = kingpin.MustParse(app.Parse(os.Args[1:]))

	return cmd
}

func Entry(v, c string) {
	version = v
	commit = c
	cmd := parseCommand()
	switch cmd.action {
	case "daemon":
		daemon()
	case "run":
		run()
	case "init":
		initialConfig(cmd.interactive)
	case "extract-web":
		err := gypsum.ExtractWebAssets(cmd.extractPath)
		if err != nil {
			fmt.Println("error when extracting: ", err)
			os.Exit(1)
		}
	case "update":
		err := gypsum.UpdateGypsum(cmd.updateVersion, cmd.githubMirror, cmd.updateForced, func(s ...interface{}) {
			fmt.Println(s...)
		})
		if err != nil {
			fmt.Println("error when updating: ", err)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown command " + cmd.action)
		os.Exit(1)
	}
}

func daemon() {
	fmt.Println("gypsum daemon started")
	for {
		cmd := exec.Command(os.Args[0], "run")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				// never mind
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if cmd.ProcessState.ExitCode() != 5 {
			os.Exit(cmd.ProcessState.ExitCode())
		}
		fmt.Println("restarting")
	}
}

func run() {
	fmt.Printf("gypsum %s, commit %s\n\n", version, commit)
	conf, err := readConfig()
	if err != nil {
		fmt.Println(err.Error())
		if runtime.GOOS == "windows" {
			fmt.Println("按回车键关闭")
			_, _ = fmt.Scanln()
		}
		os.Exit(1)
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
	//mw := io.MultiWriter(os.Stdout, logFile)
	//log.SetOutput()
	gypsum.BuildVersion = version
	gypsum.BuildCommit = commit
	gypsum.Config = &conf.Gypsum
	zero.Run(zero.Config{
		NickName:      conf.ZeroBot.NickName,
		CommandPrefix: conf.ZeroBot.CommandPrefix,
		SuperUsers:    conf.ZeroBot.SuperUsers,
		Driver: []zero.Driver{
			driver.NewWebSocketClient(conf.Host, strconv.Itoa(conf.Port), conf.AccessToken),
		},
	})
	gypsum.Init()
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		gypsum.Bot = ctx
		return false
	})
	select {}
}
