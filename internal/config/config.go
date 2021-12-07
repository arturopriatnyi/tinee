// Package config provides configuration for all application components.
package config

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

var (
	cfg  Config
	once sync.Once
)

// Config is configuration for all application components.
type Config struct {
	Service
	MongoDB
	HTTPServer
	GRPCServer
	Redis
}

// Service is configuration for service.
type Service struct {
	Domain string `envconfig:"SERVICE_DOMAIN" default:"urx.io"`
}

// MongoDB is configuration for MongoDB database.
type MongoDB struct {
	URL      string `envconfig:"MONGO_URL" default:"mongodb://localhost:27017"`
	Username string `envconfig:"MONGO_USERNAME" default:"root"`
	Password string `envconfig:"MONGO_PASSWORD" default:"password"`
	DbName   string `envconfig:"MONGO_DBNAME" default:"urx"`
}

// HTTPServer is configuration for HTTP server.
type HTTPServer struct {
	Addr string `envconfig:"HTTPSERVER_ADDR" default:":8080"`
}

// GRPCServer is configuration for gRPC server.
type GRPCServer struct {
	Addr string `envconfig:"GRPCSERVER_ADDR" default:":8081"`
}

// Redis is configuration for Redis.
type Redis struct {
	Addr     string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:"password"`
}

// Get creates Config singleton instance and returns it.
func Get() Config {
	once.Do(func() {
		if err := envconfig.Process("", &cfg); err != nil {
			zap.L().Fatal(err.Error())
		}
	})

	return cfg
}
