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

	"tinee/internal/config"
	"tinee/internal/grpc"
	"tinee/internal/http"
	"tinee/internal/mongodb"
	"tinee/internal/redis"
	"tinee/internal/service"
	"tinee/pkg/pb"
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

	rds, err := redis.Open(ctx, cfg.Redis)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	zap.L().Info("connected to Redis")

	repo := mongodb.NewLinkRepo(mgo)
	cache := redis.NewLinkCache(rds)
	s := service.New(cfg.Service, repo, cache)

	httpServer := &stdhttp.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: http.NewHandler(s),
	}

	grpcServer := stdgrpc.NewServer()
	pb.RegisterTineeURLServer(grpcServer, grpc.NewHandler(s))
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

	if err = rds.Close(); err != nil {
		zap.L().Error(err.Error())
	}
	zap.L().Info("disconnected from Redis")
}
