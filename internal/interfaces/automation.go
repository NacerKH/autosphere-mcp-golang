package interfaces

import (
	"context"
	"github.com/NacerKH/autosphere-mcp-golang/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
)

type AutomationService interface {
	LaunchJob(ctx context.Context, args models.AWXJobArgs) (models.AWXJobOutput, error)
	CheckJobStatus(ctx context.Context, args models.AWXStatusArgs) (models.AWXStatusOutput, error)
	CheckHealth(ctx context.Context, args models.HealthCheckArgs) (models.HealthCheckOutput, error)
	Autoscale(ctx context.Context, args models.AutoscaleArgs) (models.AutoscaleOutput, error)

	// New methods for enhanced AWX functionality
	ListJobs(ctx context.Context, args models.ListJobsArgs) (models.ListJobsOutput, error)
	GetJobOutput(ctx context.Context, args models.GetJobOutputArgs) (models.GetJobOutputOutput, error)
	CancelJob(ctx context.Context, args models.CancelJobArgs) (models.CancelJobOutput, error)
	ListResources(ctx context.Context, args models.ListResourcesArgs) (models.ListResourcesOutput, error)

	// Job Template management
	ListJobTemplates(ctx context.Context, args models.ListJobTemplatesArgs) (models.ListJobTemplatesOutput, error)
	CreateJobTemplate(ctx context.Context, args models.CreateJobTemplateArgs) (models.CreateJobTemplateOutput, error)
}

type HealthService interface {
	CheckComponent(component string, deep bool) models.ComponentHealth
	GetSystemMetrics() map[string]float64
	AnalyzeLoad(threshold string) string
}

type AutomationHandler interface {
	LaunchAWXJob(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	CheckAWXJobStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	CheckAutosphereHealth(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	AutoscaleAutosphere(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// New handler methods for enhanced AWX functionality
	ListAWXJobs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	GetAWXJobOutput(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	CancelAWXJob(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	ListAWXResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// Job Template management handlers
	ListJobTemplates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	CreateJobTemplate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// New observability interfaces

type ObservabilityService interface {
	QueryPrometheus(ctx context.Context, args models.QueryPrometheusArgs) (models.QueryPrometheusOutput, error)
	GetSystemMetrics(ctx context.Context, args models.GetSystemMetricsArgs) (models.GetSystemMetricsOutput, error)
	GetAlerts(ctx context.Context, args models.GetAlertsArgs) (models.GetAlertsOutput, error)
}

type ObservabilityHandler interface {
	QueryPrometheus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	GetSystemMetrics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	GetAlerts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
