package interfaces

import (
	"context"
	"github.com/NacerKH/autosphere-mcp-golang/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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
}

type HealthService interface {
	CheckComponent(component string, deep bool) models.ComponentHealth
	GetSystemMetrics() map[string]float64
	AnalyzeLoad(threshold string) string
}

type AutomationHandler interface {
	LaunchAWXJob(ctx context.Context, req *mcp.CallToolRequest, input models.AWXJobArgs) (*mcp.CallToolResult, models.AWXJobOutput, error)
	CheckAWXJobStatus(ctx context.Context, req *mcp.CallToolRequest, input models.AWXStatusArgs) (*mcp.CallToolResult, models.AWXStatusOutput, error)
	CheckAutosphereHealth(ctx context.Context, req *mcp.CallToolRequest, input models.HealthCheckArgs) (*mcp.CallToolResult, models.HealthCheckOutput, error)
	AutoscaleAutosphere(ctx context.Context, req *mcp.CallToolRequest, input models.AutoscaleArgs) (*mcp.CallToolResult, models.AutoscaleOutput, error)
	
	// New handler methods for enhanced AWX functionality
	ListAWXJobs(ctx context.Context, req *mcp.CallToolRequest, input models.ListJobsArgs) (*mcp.CallToolResult, models.ListJobsOutput, error)
	GetAWXJobOutput(ctx context.Context, req *mcp.CallToolRequest, input models.GetJobOutputArgs) (*mcp.CallToolResult, models.GetJobOutputOutput, error)
	CancelAWXJob(ctx context.Context, req *mcp.CallToolRequest, input models.CancelJobArgs) (*mcp.CallToolResult, models.CancelJobOutput, error)
	ListAWXResources(ctx context.Context, req *mcp.CallToolRequest, input models.ListResourcesArgs) (*mcp.CallToolResult, models.ListResourcesOutput, error)
}
