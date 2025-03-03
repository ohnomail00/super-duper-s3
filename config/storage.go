package config

import (
	"github.com/alecthomas/kong"
)

type Storage struct {
	Host               string   `help:"Host for the server." env:"HOST" default:"localhost"`
	Port               int      `help:"Port for the server." env:"PORT" default:"8000"`
	StoragePath        string   `help:"Path to storage files" env:"STORAGE_PATH" default:"tmp"`
	CorsAllowedOrigins []string `help:"Comma-separated list of allowed origins for CORS." env:"CORS_ALLOWED_ORIGINS" default:"*" type:"csv"`
	LogLevel           string   `help:"Log level." env:"LOG_LEVEL" default:"debug"`
}

func LoadStorageConfig() Storage {
	var cfg Storage
	kong.Parse(&cfg)
	return cfg
}
