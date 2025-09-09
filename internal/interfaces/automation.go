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
}
