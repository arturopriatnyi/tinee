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

	l, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	undo := zap.ReplaceGlobals(l)
	defer undo()

	cfg := config.Get()

	mgo, err := mongodb.Open(ctx, cfg.MongoDB)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	zap.L().Info("connected to MongoDB")

	r := mongodb.NewLinkRepo(mgo)

	s := service.New(r)

	srv := &stdhttp.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: http.NewHandler(s),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			zap.L().Fatal(err.Error())
		}
	}()
	zap.S().Infof("starting HTTP server on %s", cfg.HTTPServer.Addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error(err.Error())
	}
	zap.L().Info("HTTP server shut down gracefully")

	if err = mgo.Close(ctx); err != nil {
		zap.L().Error(err.Error())
	}
	zap.L().Info("disconnected from MongoDB")
}
