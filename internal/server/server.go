package server

import (
	"context"
	"log"
	"net/http"
	"time"
	
	"github.com/NacerKH/autosphere-mcp-golang/internal/awx"
	"github.com/NacerKH/autosphere-mcp-golang/internal/config"
	"github.com/NacerKH/autosphere-mcp-golang/internal/handlers"
	"github.com/NacerKH/autosphere-mcp-golang/internal/services"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPServer struct {
	server            *mcp.Server
	config            *config.Config
	automationHandler *handlers.AutomationHandler
}

func NewMCPServer(cfg *config.Config) *MCPServer {
	// Create AWX client with longer timeout
	awxClient := awx.NewClient(awx.ClientConfig{
		BaseURL:  cfg.AWXBaseURL,
		Username: cfg.AWXUsername,
		Password: cfg.AWXPassword,
		Token:    cfg.AWXToken,
		Timeout:  120 * time.Second, // Increased to 2 minutes
	})
	
	// Test AWX connection if credentials are provided
	if cfg.AWXUsername != "" || cfg.AWXToken != "" {
		if err := awxClient.TestConnection(context.Background()); err != nil {
			log.Printf("⚠️  AWX connection test failed: %v", err)
			log.Printf("💡 Server will still start, but AWX operations may fail")
		} else {
			log.Printf("✅ AWX connection test successful")
		}
	} else {
		log.Printf("⚠️  No AWX credentials provided. Use -awx-username/-awx-password or -awx-token flags")
	}
	
	healthService := services.NewHealthService()
	automationService := services.NewAutomationService(healthService, awxClient, cfg.AWXBaseURL)
	
	automationHandler := handlers.NewAutomationHandler(automationService)
	
	server := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.ServerName,
		Version: cfg.Version,
	}, nil)
	
	mcpServer := &MCPServer{
		server:            server,
		config:            cfg,
		automationHandler: automationHandler,
	}
	
	mcpServer.registerTools()
	
	return mcpServer
}

func (s *MCPServer) registerTools() {
	
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "launch_awx_job",
		Description: "Launch an AWX job template for Autosphere automation (deployment, scaling, health checks, backups)",
	}, s.automationHandler.LaunchAWXJob)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "check_awx_job",
		Description: "Check the status of a running or completed AWX job",
	}, s.automationHandler.CheckAWXJobStatus)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "health_check",
		Description: "Perform comprehensive health checks on Autosphere components (API, database, cache, web, workers, monitoring)",
	}, s.automationHandler.CheckAutosphereHealth)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "autoscale",
		Description: "Manage autoscaling of Autosphere services based on metrics and thresholds",
	}, s.automationHandler.AutoscaleAutosphere)
}

func (s *MCPServer) Run(ctx context.Context) error {
	s.logServerInfo()
	
	if s.config.IsHTTPMode() {
		return s.runHTTP()
	}
	return s.runSTDIO(ctx)
}

func (s *MCPServer) runHTTP() error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s.server
	}, nil)
	
	log.Printf("🌐 StreamableHTTP server starting at %s", s.config.HTTPAddr)
	log.Printf("📡 MCP endpoint: http://%s", s.config.HTTPAddr)
	log.Printf("🔧 For MCP Inspector: npx @modelcontextprotocol/inspector http://%s", s.config.HTTPAddr)
	
	return http.ListenAndServe(s.config.HTTPAddr, handler)
}

func (s *MCPServer) runSTDIO(ctx context.Context) error {
	log.Printf("📺 STDIO transport active")
	log.Printf("💡 Use -http flag to enable StreamableHTTP transport")
	
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

func (s *MCPServer) logServerInfo() {
	log.Printf("🚀 Starting Autosphere MCP server...")
	log.Printf("📡 Server: %s v%s", s.config.ServerName, s.config.Version)
	log.Printf("🛠️  Basic tools: greet, calculate, get_time")
	log.Printf("🤖 AWX/Ansible tools: launch_awx_job, check_awx_job, health_check, autoscale")
	
	if s.config.EnableDebug {
		log.Printf("🐛 Debug mode enabled")
	}
}
