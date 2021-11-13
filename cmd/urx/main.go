package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"urx/internal/config"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	cfg := config.Get()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "urx")
	})
	s := &http.Server{
		Addr: cfg.HTTPServer.Addr,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("couldn't start HTTP server")
		}
	}()
	zap.S().Infof("starting HTTP server on %s", cfg.HTTPServer.Addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		zap.L().Fatal("HTTP server couldn't shut down gracefully")
	}
	zap.L().Info("HTTP server shut down gracefully")
}
