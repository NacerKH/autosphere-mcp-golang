package config

import (
	"flag"
	"log"
)

type Config struct {
	HTTPAddr    string
	ServerName  string
	Version     string
	AWXBaseURL  string
	EnableDebug bool
}

func LoadConfig() *Config {
	httpAddr := flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")
	enableDebug := flag.Bool("debug", false, "enable debug logging")
	awxBaseURL := flag.String("awx-url", "https://awx.autosphere.local", "AWX base URL")
	
	flag.Parse()

	config := &Config{
		HTTPAddr:    *httpAddr,
		ServerName:  "autosphere-mcp-server",
		Version:     "2.0.0",
		AWXBaseURL:  *awxBaseURL,
		EnableDebug: *enableDebug,
	}

	if config.EnableDebug {
		log.Printf("Configuration loaded: %+v", config)
	}

	return config
}

func (c *Config) IsHTTPMode() bool {
	return c.HTTPAddr != ""
}
