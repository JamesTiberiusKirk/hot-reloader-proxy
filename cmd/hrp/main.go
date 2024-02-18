package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/JamesTiberiusKirk/hot-reloader-proxy/cmd/hrp/config"
	"github.com/JamesTiberiusKirk/hot-reloader-proxy/cmd/hrp/logger"
	"github.com/JamesTiberiusKirk/hot-reloader-proxy/proxy"
	"github.com/JamesTiberiusKirk/hot-reloader-proxy/watcher"
)

func main() {
	log := logger.NewDefaultLogger(true)
	c := config.GetConfig(log)

	updatedFileChan := make(chan string)
	devServerHost := fmt.Sprintf("127.0.0.1:%d", c.ProxyPort)

	var target *url.URL
	target, err := url.Parse("http://" + devServerHost)
	if err != nil {
		panic(fmt.Errorf("failed to parse proxy URL: %w", err))
	}

	log.Info("Starting proxy to host %s", target.String())
	p := proxy.New(log, c.ProxyPort, target)

	go func() {
		log.Info("Listening on %d", c.ProxyPort)
		if err := http.ListenAndServe(devServerHost, p); err != nil {
			slog.Error("Proxy failed", slog.Any("error", err))
			panic(err)
		}
	}()

	log.Info("Starting template watcher")
	go watcher.StartTemplateWatcher(log, updatedFileChan, c.WatchedFolder, c.Ignores, nil)

	for ch := range updatedFileChan {
		log.Info("File updated %s", ch)
		log.Info("Waiting for server")

	serverCheck:
		for {
			time.Sleep(100 * time.Millisecond)
			_, err := http.Get(devServerHost)
			if err != nil {
				continue
			}
			break serverCheck
		}
		log.Info("Sending reload message")

		p.SendSSE("message", "reload")
	}
}
