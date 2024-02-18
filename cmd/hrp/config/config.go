package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/JamesTiberiusKirk/hot-reloader-proxy/cmd/hrp/logger"
)

type Config struct {
	DevPort              int
	ProxyPort            int
	Ignores              []string
	Include              []string
	WatchedFolder        string
	DevServerPollingTime int
	Debug                bool
}

func GetConfig(log logger.Logger) Config {
	c := Config{}

	flag.IntVar(&c.DevPort, "dp", 5000, "port of local dev server to proxy to")
	flag.IntVar(&c.ProxyPort, "pp", 5001, "port to start the proxy on")
	flag.IntVar(&c.DevServerPollingTime, "dev-pol", 100, "dev server poling time in milliseconds")

	var ignoresRaw string
	flag.StringVar(&ignoresRaw, "ignore", "", "comma separated list strings to ignore (uses strings.Contains)")
	c.Ignores = strings.Split(ignoresRaw, ",")

	var includeRaw string
	flag.StringVar(&includeRaw, "include", "", "comma separated list strings to include (uses strings.Contains)")
	c.Include = strings.Split(includeRaw, ",")

	// TODO: need to figure out how to filter flags
	fmt.Println("args ", os.Args, len(os.Args))
	if len(os.Args) < 2 {
		log.Error("[HRP]: Need path for starting the file watcher")
		panic("need path for starting the file watcher")
	}

	c.Debug = (os.Getenv("debug") == "true")

	// TODO: then checks for env vars
	// TODO: then checks for preferences file

	return c
}
