package main

import (
	"context"
	"log"
	
	"github.com/NacerKH/autosphere-mcp-golang/internal/config"
	"github.com/NacerKH/autosphere-mcp-golang/internal/server"
)

func main() {
	cfg := config.LoadConfig()
	
	mcpServer := server.NewMCPServer(cfg)
	
	if err := mcpServer.Run(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
