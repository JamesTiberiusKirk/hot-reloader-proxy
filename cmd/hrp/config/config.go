package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	hotroloaderproxy "github.com/JamesTiberiusKirk/hot-reloader-proxy"
	"github.com/JamesTiberiusKirk/hot-reloader-proxy/cmd/hrp/logger"
	"github.com/JamesTiberiusKirk/hot-reloader-proxy/watcher"
)

type Config struct {
	DevPort              int
	ProxyPort            int
	DevServerPollingTime int
	Debug                bool
	WatcherConfig        watcher.Config
}

func GetConfig(log logger.Logger) Config {
	c := Config{}

	version := flag.Bool("version", false, "get version of hrp")
	devPort := flag.Int("dp", 5000, "port of local dev server to proxy to")
	proxyPort := flag.Int("pp", 5001, "port to start the proxy on")
	devPol := flag.Int("dev-pol", 100, "dev server poling time in milliseconds")
	ignoreRaw := flag.String("ignore", "", "comma separated list strings to ignore (uses strings.Contains)")
	includeRaw := flag.String("include", "", "comma separated list strings to include (uses strings.Contains)")
	debug := flag.Bool("debug", false, "enable debug logging")
	ignoreSuffix := flag.String("ignoreSuffix", "", "comma separated list strings to ignore (uses strings.HasSuffix)")
	includeSuffix := flag.String("includeSuffix", "", "comma separated list strings to include (uses strings.HasSuffix)")

	flag.Parse()

	if version != nil && *version {
		log.Info("Version: %s", hotroloaderproxy.Version)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) < 1 {
		log.Warn("No path supplied, using pwd")
		c.WatcherConfig.TemplateDir, _ = filepath.Abs("./")
	} else {
		c.WatcherConfig.TemplateDir = args[0]
	}

	c.DevPort = *devPort
	c.ProxyPort = *proxyPort
	c.DevServerPollingTime = *devPol
	c.Debug = *debug

	c.WatcherConfig.Ignore = strings.Split(*ignoreRaw, ",")
	c.WatcherConfig.Include = strings.Split(*includeRaw, ",")
	c.WatcherConfig.IgnoreSuffix = strings.Split(*ignoreSuffix, ",")
	c.WatcherConfig.IncludeSuffix = strings.Split(*includeSuffix, ",")

	if debug == nil {
		c.Debug = (os.Getenv("debug") == "true")
	}

	// TODO: then checks for env vars
	// TODO: then checks for preferences file

	return c
}
