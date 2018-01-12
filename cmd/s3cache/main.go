package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/zenreach/hatchet"
	"github.com/zenreach/hydroponics/internal/cache/httphandler"
	"github.com/zenreach/hydroponics/internal/cache/s3"
	"github.com/zenreach/hydroponics/internal/signals"
)

func run() int {
	logger := newLogger("")
	cfg, err := parseConfig()
	if err != nil {
		logError(logger, err, "failed to parse config")
		logger.Close()
		return 1
	}
	logger = newLogger(cfg.LogLevel)
	defer logger.Close()

	cas, err := s3.New(cfg.CASBucket, cfg.CASPrefix)
	if err != nil {
		logError(logger, err, "failed to init ac cache")
		return 1
	}

	ac, err := s3.New(cfg.ACBucket, cfg.ACPrefix)
	if err != nil {
		logError(logger, err, "failed to init ac cache")
		return 1
	}

	handler := httphandler.New(cas, ac, cfg.Timeout, logger)
	server := &http.Server{
		Addr:    cfg.Listen,
		Handler: handler,
	}

	shutdown := make(chan error, 1)
	sigs := make(chan os.Signal)
	signals.Notify(sigs)

	go func() {
		<-sigs
		logger.Log(hatchet.L{
			"message": "stop http server",
			"level":   "info",
			"address": cfg.Listen,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := server.Shutdown(ctx)
		if err != nil {
			shutdown <- err
		}
		cancel()
		close(shutdown)
	}()

	logger.Log(hatchet.L{
		"message": "configured cache",
		"level":   "debug",
		"config":  cfg,
	})

	logger.Log(hatchet.L{
		"message": "start http server",
		"level":   "info",
		"address": cfg.Listen,
	})

	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		// wait for shutdown to finish
		err = <-shutdown
	}
	if err != nil {
		logError(logger, err, "http server failure")
		return 1
	}
	return 0
}

func logError(logger hatchet.Logger, err error, msg string) {
	logger.Log(hatchet.L{
		"message": msg,
		"level":   "error",
		"error":   err,
	})
}

func main() {
	os.Exit(run())
}
