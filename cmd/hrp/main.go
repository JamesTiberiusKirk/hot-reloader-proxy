package main

import (
	"encoding/json"
	"fmt"
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

	b, _ := json.MarshalIndent(c, "", "\t")
	log.Debug("Starting with config %s", string(b))

	updatedFileChan := make(chan string)
	devServerHost := fmt.Sprintf("127.0.0.1:%d", c.DevPort)

	var target *url.URL
	target, err := url.Parse("http://" + devServerHost)
	if err != nil {
		panic(fmt.Errorf("failed to parse proxy URL: %w", err))
	}

	log.Info("Starting proxy to host %s", target.String())
	p := proxy.New(log, c.ProxyPort, target)

	go func() {
		log.Info("Listening on %d", c.ProxyPort)
		if err := http.ListenAndServe(fmt.Sprint("127.0.0.1:", c.ProxyPort), p); err != nil {
			log.Error("Proxy failed %s", err.Error())
			panic(err)
		}
	}()

	log.Info("Starting file watcher")

	go watcher.StartTemplateWatcher(log, updatedFileChan, c.WatcherConfig)

	for ch := range updatedFileChan {
		log.Info("File updated %s", ch)
		log.Info("Waiting for server")

	serverCheck:
		for {
			time.Sleep(100 * time.Millisecond)
			_, err := http.Get("http://" + devServerHost + "/ready")
			if err != nil {
				continue
			}
			break serverCheck
		}
		log.Info("Sending reload message")
		p.SendSSE("message", "reload")
	}
}
