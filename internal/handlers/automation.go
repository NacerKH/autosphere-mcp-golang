package handlers

import (
	"context"
	
	"github.com/NacerKH/autosphere-mcp-golang/internal/interfaces"
	"github.com/NacerKH/autosphere-mcp-golang/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AutomationHandler struct {
	automationService interfaces.AutomationService
}

func NewAutomationHandler(automationService interfaces.AutomationService) *AutomationHandler {
	return &AutomationHandler{
		automationService: automationService,
	}
}

func (h *AutomationHandler) LaunchAWXJob(ctx context.Context, req *mcp.CallToolRequest, input models.AWXJobArgs) (*mcp.CallToolResult, models.AWXJobOutput, error) {
	output, err := h.automationService.LaunchJob(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.AWXJobOutput{}, err
	}
	return nil, output, nil
}

func (h *AutomationHandler) CheckAWXJobStatus(ctx context.Context, req *mcp.CallToolRequest, input models.AWXStatusArgs) (*mcp.CallToolResult, models.AWXStatusOutput, error) {
	output, err := h.automationService.CheckJobStatus(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.AWXStatusOutput{}, err
	}
	return nil, output, nil
}

func (h *AutomationHandler) CheckAutosphereHealth(ctx context.Context, req *mcp.CallToolRequest, input models.HealthCheckArgs) (*mcp.CallToolResult, models.HealthCheckOutput, error) {
	output, err := h.automationService.CheckHealth(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.HealthCheckOutput{}, err
	}
	return nil, output, nil
}

func (h *AutomationHandler) AutoscaleAutosphere(ctx context.Context, req *mcp.CallToolRequest, input models.AutoscaleArgs) (*mcp.CallToolResult, models.AutoscaleOutput, error) {
	output, err := h.automationService.Autoscale(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.AutoscaleOutput{}, err
	}
	return nil, output, nil
}

// New handler methods for additional tools

func (h *AutomationHandler) ListAWXJobs(ctx context.Context, req *mcp.CallToolRequest, input models.ListJobsArgs) (*mcp.CallToolResult, models.ListJobsOutput, error) {
	output, err := h.automationService.ListJobs(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.ListJobsOutput{}, err
	}
	return nil, output, nil
}

func (h *AutomationHandler) GetAWXJobOutput(ctx context.Context, req *mcp.CallToolRequest, input models.GetJobOutputArgs) (*mcp.CallToolResult, models.GetJobOutputOutput, error) {
	output, err := h.automationService.GetJobOutput(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.GetJobOutputOutput{}, err
	}
	return nil, output, nil
}

func (h *AutomationHandler) CancelAWXJob(ctx context.Context, req *mcp.CallToolRequest, input models.CancelJobArgs) (*mcp.CallToolResult, models.CancelJobOutput, error) {
	output, err := h.automationService.CancelJob(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.CancelJobOutput{}, err
	}
	return nil, output, nil
}

func (h *AutomationHandler) ListAWXResources(ctx context.Context, req *mcp.CallToolRequest, input models.ListResourcesArgs) (*mcp.CallToolResult, models.ListResourcesOutput, error) {
	output, err := h.automationService.ListResources(ctx, input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, models.ListResourcesOutput{}, err
	}
	return nil, output, nil
}
