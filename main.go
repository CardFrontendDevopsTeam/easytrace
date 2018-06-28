package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/weAutomateEverything/go2hal/remoteTelegramCommands"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/CardFrontendDevopsTeam/easytrace/easytraceCache"
	)

func main() {

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowAll())
	logger = log.With(logger, "ts", log.DefaultTimestamp)
	remoteTelegramService := remoteTelegramCommands.NewRemoteCommandClientService()
	easytraceCache.NewService(remoteTelegramService)
	http.Handle("/metrics", promhttp.Handler())
	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", ":8001", "msg", "listening")
		errs <- http.ListenAndServe(":8001", nil)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}
func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
