package main

import (
	"context"
	"log"
	"net"
	stdhttp "net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
	stdgrpc "google.golang.org/grpc"

	"urx/internal/config"
	"urx/internal/grpc"
	"urx/internal/http"
	"urx/internal/mongodb"
	"urx/internal/service"
	"urx/pkg/pb"
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

	r := mongodb.NewLinkRepo(mgo)

	s := service.New(cfg.Service, r)

	httpServer := &stdhttp.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: http.NewHandler(s),
	}

	grpcServer := stdgrpc.NewServer()
	pb.RegisterURXServer(grpcServer, grpc.NewHandler(s))
	l, err := net.Listen("tcp", cfg.GRPCServer.Addr)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			zap.L().Fatal(err.Error())
		}
	}()
	zap.S().Infof("starting HTTP server on %s", cfg.HTTPServer.Addr)

	go func() {
		if err := grpcServer.Serve(l); err != nil {
			zap.L().Fatal(err.Error())
		}
	}()
	zap.S().Infof("starting gRPC server on %s", cfg.GRPCServer.Addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		zap.L().Fatal(err.Error())
	}
	zap.L().Info("HTTP server shut down gracefully")

	grpcServer.GracefulStop()
	zap.L().Info("gRPC server shut down gracefully")

	if err = mgo.Close(ctx); err != nil {
		zap.L().Error(err.Error())
	}
	zap.L().Info("disconnected from MongoDB")
}
