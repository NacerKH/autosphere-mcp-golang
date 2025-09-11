package server

import (
	"context"
	"log"
	"time"
	
	"github.com/NacerKH/autosphere-mcp-golang/internal/awx"
	"github.com/NacerKH/autosphere-mcp-golang/internal/config"
	"github.com/NacerKH/autosphere-mcp-golang/internal/handlers"
	"github.com/NacerKH/autosphere-mcp-golang/internal/handlers/prompts"
	"github.com/NacerKH/autosphere-mcp-golang/internal/handlers/resources"
	"github.com/NacerKH/autosphere-mcp-golang/internal/services"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	server              *server.MCPServer
	config              *config.Config
	automationHandler   *handlers.AutomationHandler
	observabilityHandler *handlers.ObservabilityHandler
	resourceHandler     *resources.ResourceHandler
	promptsHandler      *prompts.PromptsHandler
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
	resourceHandler := resources.NewResourceHandler()
	promptsHandler := prompts.NewPromptsHandler()

	
	server := server.NewMCPServer(
		cfg.ServerName,
		cfg.Version,
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
	)
	
	mcpServer := &MCPServer{
		server:            server,
		config:            cfg,
		automationHandler: automationHandler,
		resourceHandler:   resourceHandler,
		promptsHandler:    promptsHandler,
	}
	
	mcpServer.registerTools()
	mcpServer.registerResources()
	mcpServer.registerPrompts()
	
	return mcpServer
}

func (s *MCPServer) registerTools() {
	// Launch AWX Job Tool
	launchAWXTool := mcp.NewTool("launch_awx_job",
		mcp.WithDescription("Launch an AWX job template for Autosphere automation (deployment, scaling, health checks, backups)"),
		mcp.WithString("job_template", mcp.Required(), mcp.Description("The name or ID of the AWX job template")),
		mcp.WithString("extra_vars", mcp.Description("Extra variables to pass to the job (JSON string)")),
		mcp.WithString("inventory", mcp.Description("Inventory name or ID (optional)")),
		mcp.WithString("limit", mcp.Description("Limit the job to specific hosts (optional)")),
		mcp.WithString("tags", mcp.Description("Ansible tags to run (optional)")),
		mcp.WithString("skip_tags", mcp.Description("Ansible tags to skip (optional)")),
	)
	s.server.AddTool(launchAWXTool, s.automationHandler.LaunchAWXJob)

	// Check AWX Job Status Tool
	checkAWXTool := mcp.NewTool("check_awx_job",
		mcp.WithDescription("Check the status of a running or completed AWX job"),
		mcp.WithString("job_id", mcp.Required(), mcp.Description("The AWX job ID to check")),
	)
	s.server.AddTool(checkAWXTool, s.automationHandler.CheckAWXJobStatus)

	// Health Check Tool
	healthCheckTool := mcp.NewTool("health_check",
		mcp.WithDescription("Perform comprehensive health checks on Autosphere components (API, database, cache, web, workers, monitoring)"),
		mcp.WithString("component", mcp.Description("Specific component to check (api, database, cache, all)")),
		mcp.WithString("deep", mcp.Description("Perform deep health checks (true/false)")),
	)
	s.server.AddTool(healthCheckTool, s.automationHandler.CheckAutosphereHealth)

	// Autoscale Tool
	autoscaleTool := mcp.NewTool("autoscale",
		mcp.WithDescription("Manage autoscaling of Autosphere services based on metrics and thresholds"),
		mcp.WithString("action", mcp.Required(), mcp.Description("Autoscaling action (scale_up, scale_down, analyze, auto)")),
		mcp.WithString("service", mcp.Description("Specific service to scale (optional)")),
		mcp.WithString("replicas", mcp.Description("Target number of replicas (for manual scaling)")),
		mcp.WithString("threshold", mcp.Description("Scaling threshold (cpu_high, memory_high, load_high)")),
	)
	s.server.AddTool(autoscaleTool, s.automationHandler.AutoscaleAutosphere)

	// List AWX Jobs Tool
	listJobsTool := mcp.NewTool("list_awx_jobs",
		mcp.WithDescription("List AWX jobs with optional filtering by status and limit"),
		mcp.WithString("limit", mcp.Description("Maximum number of jobs to return (default: 20)")),
		mcp.WithString("status", mcp.Description("Filter by job status (successful, failed, running, pending)")),
	)
	s.server.AddTool(listJobsTool, s.automationHandler.ListAWXJobs)

	// Get AWX Job Output Tool
	getJobOutputTool := mcp.NewTool("get_job_output",
		mcp.WithDescription("Get the output/logs of a specific AWX job"),
		mcp.WithString("job_id", mcp.Required(), mcp.Description("The AWX job ID to get output for")),
	)
	s.server.AddTool(getJobOutputTool, s.automationHandler.GetAWXJobOutput)

	// Cancel AWX Job Tool
	cancelJobTool := mcp.NewTool("cancel_awx_job",
		mcp.WithDescription("Cancel a running AWX job"),
		mcp.WithString("job_id", mcp.Required(), mcp.Description("The AWX job ID to cancel")),
	)
	s.server.AddTool(cancelJobTool, s.automationHandler.CancelAWXJob)

	// List AWX Resources Tool
	listResourcesTool := mcp.NewTool("list_awx_resources",
		mcp.WithDescription("List AWX resources (job templates, inventories, projects)"),
		mcp.WithString("resource_type", mcp.Required(), mcp.Description("Type of resource (templates, inventories, projects)")),
	)
	s.server.AddTool(listResourcesTool, s.automationHandler.ListAWXResources)
}

func (s *MCPServer) registerResources() {
	// Register system configuration resource
	configResource := mcp.NewResource(
		"autosphere://config",
		"Autosphere Configuration",
		mcp.WithResourceDescription("Complete Autosphere system settings and configuration"),
		mcp.WithMIMEType("application/json"),
	)
	s.server.AddResource(configResource, s.resourceHandler.GetAutosphereConfig)

	// Register deployment manifest resource
	manifestResource := mcp.NewResource(
		"autosphere://deployment-manifest",
		"Deployment Manifest",
		mcp.WithResourceDescription("Kubernetes deployment manifest for Autosphere"),
		mcp.WithMIMEType("text/yaml"),
	)
	s.server.AddResource(manifestResource, s.resourceHandler.GetDeploymentManifest)

	// Register health report resource
	healthResource := mcp.NewResource(
		"autosphere://health-report",
		"Health Check Report",
		mcp.WithResourceDescription("Real-time system health status and metrics"),
		mcp.WithMIMEType("application/json"),
	)
	s.server.AddResource(healthResource, s.resourceHandler.GetHealthCheckReport)

	// Register AWX templates resource
	awxResource := mcp.NewResource(
		"autosphere://awx-templates",
		"AWX Job Templates",
		mcp.WithResourceDescription("Available AWX job templates for automation"),
		mcp.WithMIMEType("application/json"),
	)
	s.server.AddResource(awxResource, s.resourceHandler.GetAWXJobTemplates)
}

func (s *MCPServer) registerPrompts() {
	// Register deployment planning prompt
	deploymentPrompt := mcp.NewPrompt(
		"deployment_planning",
		mcp.WithPromptDescription("Comprehensive deployment planning and guidance"),
		mcp.WithArgument("environment", mcp.RequiredArgument(), mcp.ArgumentDescription("Target environment (production, staging, development)")),
		mcp.WithArgument("version", mcp.RequiredArgument(), mcp.ArgumentDescription("Application version to deploy")),
		mcp.WithArgument("components", mcp.ArgumentDescription("Specific components to deploy (optional)")),
	)
	s.server.AddPrompt(deploymentPrompt, s.promptsHandler.DeploymentPlanning)

	// Register troubleshooting prompt
	troubleshootingPrompt := mcp.NewPrompt(
		"troubleshooting",
		mcp.WithPromptDescription("Systematic troubleshooting guidance for issues"),
		mcp.WithArgument("issue", mcp.RequiredArgument(), mcp.ArgumentDescription("Brief description of the problem")),
		mcp.WithArgument("component", mcp.RequiredArgument(), mcp.ArgumentDescription("Affected component (api, database, cache, etc.)")),
		mcp.WithArgument("symptoms", mcp.ArgumentDescription("Observed symptoms or error messages (optional)")),
	)
	s.server.AddPrompt(troubleshootingPrompt, s.promptsHandler.TroubleshootingGuide)
}



func (s *MCPServer) Run(ctx context.Context) error {
	s.logServerInfo()
	
	if s.config.IsHTTPMode() {
		return s.runHTTP()
	}
	return s.runSTDIO(ctx)
}

func (s *MCPServer) runHTTP() error {
	// For HTTP mode, we'll use StreamableHTTP server from mcp-go
	log.Printf("🌐 StreamableHTTP server starting at %s", s.config.HTTPAddr)
	log.Printf("📡 MCP endpoint: http://%s/mcp", s.config.HTTPAddr)
	
	// Create StreamableHTTP server
	streamableServer := server.NewStreamableHTTPServer(s.server)
	
	return streamableServer.Start(s.config.HTTPAddr)
}

func (s *MCPServer) runSTDIO(ctx context.Context) error {
	log.Printf("📺 STDIO transport active")
	log.Printf("💡 Use -http flag to enable HTTP transport")
	
	return server.ServeStdio(s.server)
}

func (s *MCPServer) logServerInfo() {
	log.Printf("Starting Autosphere MCP server...")
	log.Printf("Server: %s v%s", s.config.ServerName, s.config.Version)
	log.Printf("Core AWX tools: launch_awx_job, check_awx_job, health_check, autoscale")
	log.Printf("Enhanced AWX tools: list_awx_jobs, get_job_output, cancel_awx_job, list_awx_resources")
	log.Printf("Resources: autosphere://config, autosphere://deployment-manifest, autosphere://health-report, autosphere://awx-templates")
	log.Printf("Prompts: deployment_planning, troubleshooting, scaling_decision, incident_response")
	
	if s.config.EnableDebug {
		log.Printf("Debug mode enabled")
	}
}
