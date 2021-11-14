package main

import (
	"context"
	"log"
	stdhttp "net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"urx/internal/config"
	"urx/internal/http"
	"urx/internal/mongodb"
	"urx/internal/service"
)

func main() {
	ctx := context.Background()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	cfg := config.Get()

	mgo, err := mongodb.Open(ctx, cfg.MongoDB)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	zap.L().Info("connected to MongoDB")

	svc := service.New(mongodb.NewLinkRepo(mgo))

	s := &stdhttp.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: http.NewHandler(svc),
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
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

	if err = mgo.Close(ctx); err != nil {
		zap.L().Error("failed to disconnect from MongoDB")
	}
	zap.L().Info("disconnected from MongoDB")
}
