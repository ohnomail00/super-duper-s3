package config

import (
	"github.com/alecthomas/kong"
)

type Gateway struct {
	Host               string   `help:"Host for the server." env:"HOST" default:"localhost"`
	Port               int      `help:"Port for the server." env:"PORT" default:"8000"`
	StorageAddrs       []string `help:"Comma-separated list of storage addresses." env:"STORAGE_ADDRS" default:"http://storage1.example.com,http://storage2.example.com,http://storage3.example.com,http://storage4.example.com,http://storage5.example.com,http://storage6.example.com" type:"csv"`
	PartCount          int      `help:"Number of parts for objects to split." env:"PART_COUNT" default:"6"`
	VirtualReplicas    int      `help:"Number of replicas for virtual nodes." env:"VIRTUAL_REPLICAS" default:"100"`
	CorsAllowedOrigins []string `help:"Comma-separated list of allowed origins for CORS." env:"CORS_ALLOWED_ORIGINS" default:"*" type:"csv"`
	LogLevel           string   `help:"Log level." env:"LOG_LEVEL" default:"debug"`
}

func LoadGatewayConfig() Gateway {
	var cfg Gateway
	kong.Parse(&cfg)
	return cfg
}
