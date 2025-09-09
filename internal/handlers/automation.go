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
