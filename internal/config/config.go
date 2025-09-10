package config

import (
	"flag"
	"log"
)

type Config struct {
	HTTPAddr     string
	ServerName   string
	Version      string
	AWXBaseURL   string
	AWXUsername  string
	AWXPassword  string
	AWXToken     string
	EnableDebug  bool
}

func LoadConfig() *Config {
	httpAddr := flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")
	enableDebug := flag.Bool("debug", false, "enable debug logging")
	awxBaseURL := flag.String("awx-url", "http://awx.autosphere.local:30930", "AWX base URL")
	awxUsername := flag.String("awx-username", "", "AWX username")
	awxPassword := flag.String("awx-password", "", "AWX password")
	awxToken := flag.String("awx-token", "", "AWX API token (alternative to username/password)")
	
	flag.Parse()

	config := &Config{
		HTTPAddr:     *httpAddr,
		ServerName:   "autosphere-mcp-server",
		Version:      "2.0.0",
		AWXBaseURL:   *awxBaseURL,
		AWXUsername:  *awxUsername,
		AWXPassword:  *awxPassword,
		AWXToken:     *awxToken,
		EnableDebug:  *enableDebug,
	}

	if config.EnableDebug {
		log.Printf("Configuration loaded: %+v", config)
	}

	return config
}

func (c *Config) IsHTTPMode() bool {
	return c.HTTPAddr != ""
}
