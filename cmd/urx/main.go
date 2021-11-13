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
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "urx")
	})
	s := &http.Server{
		Addr: ":8080",
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("couldn't start HTTP server")
		}
	}()
	zap.L().Info("starting HTTP server on :8080")

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
