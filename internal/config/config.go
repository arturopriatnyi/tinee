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
	HTTPServer
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