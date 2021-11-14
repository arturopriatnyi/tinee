// Package config provides configuration for all project components.
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

// Config is configuration for all project components.
type Config struct {
	Service
	MongoDB
	HTTPServer
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

// Get creates Config singleton instance and returns it.
func Get() Config {
	once.Do(func() {
		if err := envconfig.Process("", &cfg); err != nil {
			zap.L().Fatal(err.Error())
		}
	})

	return cfg
}
