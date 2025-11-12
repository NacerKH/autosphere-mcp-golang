package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/NacerKH/autosphere-mcp-golang/internal/interfaces"
	"github.com/NacerKH/autosphere-mcp-golang/internal/models"
	"github.com/mark3labs/mcp-go/mcp"
)

type AutomationHandler struct {
	automationService interfaces.AutomationService
}

func NewAutomationHandler(automationService interfaces.AutomationService) *AutomationHandler {
	return &AutomationHandler{
		automationService: automationService,
	}
}

// LaunchAWXJob launches an AWX job template with provided parameters
func (h *AutomationHandler) LaunchAWXJob(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse arguments using mcp-go helper methods
	args := models.AWXJobArgs{}
	
	// Required: job_template
	jobTemplate, err := request.RequireString("job_template")
	if err != nil {
		return mcp.NewToolResultError("job_template is required"), nil
	}
	args.JobTemplate = jobTemplate
	
	// Optional: extra_vars (parse as JSON string)
	extraVarsStr := request.GetString("extra_vars", "")
	if extraVarsStr != "" {
		extraVars := make(map[string]string)
		if err := json.Unmarshal([]byte(extraVarsStr), &extraVars); err != nil {
			log.Printf("Warning: Failed to parse extra_vars: %v", err)
		} else {
			args.ExtraVars = extraVars
		}
	}
	
	// Optional parameters using GetString with defaults
	args.Inventory = request.GetString("inventory", "")
	args.Limit = request.GetString("limit", "")
	args.Tags = request.GetString("tags", "")
	args.SkipTags = request.GetString("skip_tags", "")
	
	// Call the actual automation service
	output, err := h.automationService.LaunchJob(ctx, args)
	if err != nil {
		log.Printf("AWX job launch failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to launch AWX job: %v", err)), nil
	}
	
	// Format successful response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	message := fmt.Sprintf("✅ AWX Job Launched Successfully\n\n**Job Details:**\n- Job ID: %d\n- Status: %s\n- AWX URL: %s\n\n**Full Response:**\n```json\n%s\n```", 
		output.JobID, output.Status, output.URL, string(resultJSON))
	
	return mcp.NewToolResultText(message), nil
}

// CheckAWXJobStatus checks the status of a running or completed AWX job
func (h *AutomationHandler) CheckAWXJobStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.AWXStatusArgs{}
	
	// Required: job_id
	jobIDStr, err := request.RequireString("job_id")
	if err != nil {
		return mcp.NewToolResultError("job_id is required"), nil
	}
	
	if jobID, err := strconv.Atoi(jobIDStr); err == nil {
		args.JobID = jobID
	} else {
		return mcp.NewToolResultError("job_id must be a valid integer"), nil
	}
	
	// Call the automation service
	output, err := h.automationService.CheckJobStatus(ctx, args)
	if err != nil {
		log.Printf("AWX job status check failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to check AWX job status: %v", err)), nil
	}
	
	// Format response with status information
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	
	statusEmoji := "🔄"
	switch output.Status {
	case "successful":
		statusEmoji = "✅"
	case "failed":
		statusEmoji = "❌"
	case "running":
		statusEmoji = "🔄"
	case "pending":
		statusEmoji = "⏳"
	}
	
	message := fmt.Sprintf("%s AWX Job Status\n\n**Job %d Status: %s**\n\n📅 **Timeline:**\n- Started: %s\n- Elapsed: %s\n", 
		statusEmoji, output.JobID, output.Status, output.StartedAt, output.ElapsedTime)
	
	if output.FinishedAt != "" {
		message += fmt.Sprintf("- Finished: %s\n", output.FinishedAt)
	}
	
	message += fmt.Sprintf("\n**Full Response:**\n```json\n%s\n```", string(resultJSON))
	
	return mcp.NewToolResultText(message), nil
}

// CheckAutosphereHealth performs comprehensive health checks on Autosphere components
func (h *AutomationHandler) CheckAutosphereHealth(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.HealthCheckArgs{}
	
	// Optional parameters using GetString with defaults
	args.Component = request.GetString("component", "")
	deepStr := request.GetString("deep", "false")
	args.Deep = deepStr == "true"
	
	// Call the automation service
	output, err := h.automationService.CheckHealth(ctx, args)
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to perform health check: %v", err)), nil
	}
	
	// Format health check response
	statusEmoji := "✅"
	switch output.OverallStatus {
	case "healthy":
		statusEmoji = "✅"
	case "warning":
		statusEmoji = "⚠️"
	case "critical":
		statusEmoji = "❌"
	default:
		statusEmoji = "❓"
	}
	
	message := fmt.Sprintf("%s Autosphere Health Check\n\n**Overall Status: %s**\n📅 Checked at: %s\n\n", 
		statusEmoji, output.OverallStatus, output.Timestamp)
	
	// Add component details
	message += "**Component Status:**\n"
	for name, component := range output.Components {
		componentEmoji := "✅"
		switch component.Status {
		case "healthy":
			componentEmoji = "✅"
		case "warning":
			componentEmoji = "⚠️"
		case "critical":
			componentEmoji = "❌"
		default:
			componentEmoji = "❓"
		}
		message += fmt.Sprintf("- %s **%s**: %s - %s\n", componentEmoji, name, component.Status, component.Details)
	}
	
	// Add recommendations if any
	if len(output.Recommendations) > 0 {
		message += "\n**💡 Recommendations:**\n"
		for _, rec := range output.Recommendations {
			message += fmt.Sprintf("- %s\n", rec)
		}
	}
	
	// Add full JSON response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	message += fmt.Sprintf("\n**Full Response:**\n```json\n%s\n```", string(resultJSON))
	
	return mcp.NewToolResultText(message), nil
}

// AutoscaleAutosphere manages autoscaling of Autosphere services based on metrics and thresholds
func (h *AutomationHandler) AutoscaleAutosphere(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.AutoscaleArgs{}
	
	// Required: action
	action, err := request.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError("action is required (scale_up, scale_down, analyze, auto)"), nil
	}
	args.Action = action
	
	// Optional parameters using GetString with defaults
	args.Service = request.GetString("service", "")
	args.Threshold = request.GetString("threshold", "")
	
	// Parse replicas if provided
	replicasStr := request.GetString("replicas", "0")
	if replicas, err := strconv.Atoi(replicasStr); err == nil {
		args.Replicas = replicas
	}
	
	// Call the automation service
	output, err := h.automationService.Autoscale(ctx, args)
	if err != nil {
		log.Printf("Autoscaling failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to perform autoscaling: %v", err)), nil
	}
	
	// Format autoscale response
	actionEmoji := "📈"
	switch output.Action {
	case "scale_up":
		actionEmoji = "📈"
	case "scale_down":
		actionEmoji = "📉"
	case "analyze":
		actionEmoji = "🔍"
	case "auto":
		actionEmoji = "🤖"
	}
	
	message := fmt.Sprintf("%s Autoscaling Action: %s\n\n**Service:** %s\n**Scaling:** %d → %d replicas\n**Reason:** %s\n**Status:** %s\n", 
		actionEmoji, output.Action, output.Service, output.OldReplicas, output.NewReplicas, output.Reason, output.Status)
	
	if output.JobID > 0 {
		message += fmt.Sprintf("**AWX Job ID:** %d\n", output.JobID)
	}
	
	// Add full JSON response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	message += fmt.Sprintf("\n**Full Response:**\n```json\n%s\n```", string(resultJSON))
	
	return mcp.NewToolResultText(message), nil
}

// ListAWXJobs lists AWX jobs with optional filtering
func (h *AutomationHandler) ListAWXJobs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.ListJobsArgs{
		Limit: 20, // Default limit
	}
	
	// Optional parameters using GetString with defaults
	limitStr := request.GetString("limit", "20")
	if limit, err := strconv.Atoi(limitStr); err == nil {
		args.Limit = limit
	}
	
	args.Status = request.GetString("status", "")
	
	// Call the automation service
	output, err := h.automationService.ListJobs(ctx, args)
	if err != nil {
		log.Printf("List AWX jobs failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list AWX jobs: %v", err)), nil
	}
	
	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📋 AWX Jobs List\n\n**Found %d jobs (showing %d):**\n\n", output.Total, len(output.Jobs)))

	for _, job := range output.Jobs {
		statusEmoji := "🔄"
		switch job.Status {
		case "successful":
			statusEmoji = "✅"
		case "failed":
			statusEmoji = "❌"
		case "running":
			statusEmoji = "🔄"
		case "pending":
			statusEmoji = "⏳"
		}

		builder.WriteString(fmt.Sprintf("%s **Job %d**: %s\n", statusEmoji, job.ID, job.Name))
		builder.WriteString(fmt.Sprintf("   - Template: %s\n", job.Template))
		builder.WriteString(fmt.Sprintf("   - Status: %s\n", job.Status))
		if job.ElapsedTime != "" {
			builder.WriteString(fmt.Sprintf("   - Duration: %s\n", job.ElapsedTime))
		}
		builder.WriteString("\n")
	}

	// Add full JSON response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	builder.WriteString(fmt.Sprintf("**Full Response:**\n```json\n%s\n```", string(resultJSON)))
	message := builder.String()
	
	return mcp.NewToolResultText(message), nil
}

// GetAWXJobOutput gets the output/logs of a specific AWX job
func (h *AutomationHandler) GetAWXJobOutput(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.GetJobOutputArgs{}
	
	// Required: job_id
	jobIDStr, err := request.RequireString("job_id")
	if err != nil {
		return mcp.NewToolResultError("job_id is required"), nil
	}
	
	if jobID, err := strconv.Atoi(jobIDStr); err == nil {
		args.JobID = jobID
	} else {
		return mcp.NewToolResultError("job_id must be a valid integer"), nil
	}
	
	// Call the automation service
	output, err := h.automationService.GetJobOutput(ctx, args)
	if err != nil {
		log.Printf("Get AWX job output failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get AWX job output: %v", err)), nil
	}
	
	// Format job output response
	message := fmt.Sprintf("📜 AWX Job %d Output\n\n**Job Logs:**\n```\n%s\n```", output.JobID, output.Output)
	
	return mcp.NewToolResultText(message), nil
}

// CancelAWXJob cancels a running AWX job
func (h *AutomationHandler) CancelAWXJob(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.CancelJobArgs{}
	
	// Required: job_id
	jobIDStr, err := request.RequireString("job_id")
	if err != nil {
		return mcp.NewToolResultError("job_id is required"), nil
	}
	
	if jobID, err := strconv.Atoi(jobIDStr); err == nil {
		args.JobID = jobID
	} else {
		return mcp.NewToolResultError("job_id must be a valid integer"), nil
	}
	
	// Call the automation service
	output, err := h.automationService.CancelJob(ctx, args)
	if err != nil {
		log.Printf("Cancel AWX job failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to cancel AWX job: %v", err)), nil
	}
	
	// Format cancellation response
	statusEmoji := "✅"
	if output.Status != "canceled" {
		statusEmoji = "⚠️"
	}
	
	message := fmt.Sprintf("%s AWX Job Cancellation\n\n**Job %d**: %s\n**Status:** %s\n**Message:** %s", 
		statusEmoji, output.JobID, output.Status, output.Status, output.Message)
	
	return mcp.NewToolResultText(message), nil
}

// ListAWXResources lists AWX resources (job templates, inventories, projects)
func (h *AutomationHandler) ListAWXResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.ListResourcesArgs{}
	
	// Required: resource_type
	resourceType, err := request.RequireString("resource_type")
	if err != nil {
		return mcp.NewToolResultError("resource_type is required (templates, inventories, projects)"), nil
	}
	args.ResourceType = resourceType
	
	// Call the automation service
	output, err := h.automationService.ListResources(ctx, args)
	if err != nil {
		log.Printf("List AWX resources failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list AWX resources: %v", err)), nil
	}
	
	// Format resource list response
	typeEmoji := "📋"
	switch output.ResourceType {
	case "templates":
		typeEmoji = "🔧"
	case "inventories":
		typeEmoji = "📊"
	case "projects":
		typeEmoji = "📁"
	}
	
	message := fmt.Sprintf("%s AWX %s\n\n**Found %d %s:**\n\n", typeEmoji, output.ResourceType, output.Total, output.ResourceType)
	
	// Add resource list (simplified view)
	for i, resource := range output.Resources {
		if i >= 10 { // Limit display to first 10 for readability
			message += fmt.Sprintf("... and %d more\n", output.Total-10)
			break
		}
		
		// Try to extract basic info from interface{}
		if resourceMap, ok := resource.(map[string]interface{}); ok {
			name := "Unknown"
			id := 0
			
			if nameVal, exists := resourceMap["name"]; exists {
				if nameStr, ok := nameVal.(string); ok {
					name = nameStr
				}
			}
			if idVal, exists := resourceMap["id"]; exists {
				if idFloat, ok := idVal.(float64); ok {
					id = int(idFloat)
				}
			}
			
			message += fmt.Sprintf("- **%s** (ID: %d)\n", name, id)
		}
	}

	// Add full JSON response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	message += fmt.Sprintf("\n**Full Response:**\n```json\n%s\n```", string(resultJSON))

	return mcp.NewToolResultText(message), nil
}

// ListJobTemplates lists all AWX job templates
func (h *AutomationHandler) ListJobTemplates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.ListJobTemplatesArgs{}

	// Call the automation service
	output, err := h.automationService.ListJobTemplates(ctx, args)
	if err != nil {
		log.Printf("List job templates failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list job templates: %v", err)), nil
	}

	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("🔧 AWX Job Templates\n\n**Found %d job templates:**\n\n", output.Total))

	for _, template := range output.Templates {
		builder.WriteString(fmt.Sprintf("**%s** (ID: %d)\n", template.Name, template.ID))
		if template.Description != "" {
			builder.WriteString(fmt.Sprintf("   - Description: %s\n", template.Description))
		}
		builder.WriteString(fmt.Sprintf("   - Playbook: %s\n", template.Playbook))
		builder.WriteString(fmt.Sprintf("   - Project ID: %d\n", template.Project))
		builder.WriteString(fmt.Sprintf("   - Inventory ID: %d\n", template.Inventory))
		builder.WriteString("\n")
	}

	// Add full JSON response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	builder.WriteString(fmt.Sprintf("**Full Response:**\n```json\n%s\n```", string(resultJSON)))
	message := builder.String()

	return mcp.NewToolResultText(message), nil
}

// CreateJobTemplate creates a new AWX job template
func (h *AutomationHandler) CreateJobTemplate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.CreateJobTemplateArgs{}

	// Required: name
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}
	args.Name = name

	// Required: inventory (as integer)
	inventoryStr, err := request.RequireString("inventory")
	if err != nil {
		return mcp.NewToolResultError("inventory ID is required"), nil
	}
	if inventory, err := strconv.Atoi(inventoryStr); err == nil {
		args.Inventory = inventory
	} else {
		return mcp.NewToolResultError("inventory must be a valid integer ID"), nil
	}

	// Required: project (as integer)
	projectStr, err := request.RequireString("project")
	if err != nil {
		return mcp.NewToolResultError("project ID is required"), nil
	}
	if project, err := strconv.Atoi(projectStr); err == nil {
		args.Project = project
	} else {
		return mcp.NewToolResultError("project must be a valid integer ID"), nil
	}

	// Required: playbook
	playbook, err := request.RequireString("playbook")
	if err != nil {
		return mcp.NewToolResultError("playbook path is required"), nil
	}
	args.Playbook = playbook

	// Optional parameters
	args.Description = request.GetString("description", "")
	args.JobType = request.GetString("job_type", "run")

	verbosityStr := request.GetString("verbosity", "0")
	if verbosity, err := strconv.Atoi(verbosityStr); err == nil {
		args.Verbosity = verbosity
	}

	// Call the automation service
	output, err := h.automationService.CreateJobTemplate(ctx, args)
	if err != nil {
		log.Printf("Create job template failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create job template: %v", err)), nil
	}

	// Format creation response
	message := fmt.Sprintf("✅ Job Template Created Successfully\n\n**Template Details:**\n- ID: %d\n- Name: %s\n- Description: %s\n- Status: %s\n\n**Message:** %s\n",
		output.ID, output.Name, output.Description, output.Status, output.Message)

	// Add full JSON response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	message += fmt.Sprintf("\n**Full Response:**\n```json\n%s\n```", string(resultJSON))

	return mcp.NewToolResultText(message), nil
}

// GetCacheStats retrieves cache statistics
func (h *AutomationHandler) GetCacheStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := models.GetCacheStatsArgs{}

	// Call the automation service
	output, err := h.automationService.GetCacheStats(ctx, args)
	if err != nil {
		log.Printf("Get cache stats failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get cache stats: %v", err)), nil
	}

	// Format cache stats response
	var builder strings.Builder
	builder.WriteString("📊 Cache Performance Statistics\n\n")
	builder.WriteString(fmt.Sprintf("**%s**\n\n", output.Summary))

	builder.WriteString("**AWX Client Cache:**\n")
	builder.WriteString(fmt.Sprintf("  • Hits: %d\n", output.AWXCache.Hits))
	builder.WriteString(fmt.Sprintf("  • Misses: %d\n", output.AWXCache.Misses))
	builder.WriteString(fmt.Sprintf("  • Hit Rate: %.1f%%\n", output.AWXCache.HitRate))
	builder.WriteString(fmt.Sprintf("  • Cached Items: %d\n", output.AWXCache.CurrentSize))
	builder.WriteString(fmt.Sprintf("  • Total Sets: %d\n", output.AWXCache.Sets))
	builder.WriteString(fmt.Sprintf("  • Evictions: %d\n", output.AWXCache.Evictions))

	// Performance interpretation
	builder.WriteString("\n**Performance Impact:**\n")
	if output.AWXCache.HitRate >= 70 {
		builder.WriteString("  ✅ Excellent cache performance - 70%+ of requests served from cache\n")
	} else if output.AWXCache.HitRate >= 50 {
		builder.WriteString("  ⚠️  Good cache performance - consider increasing TTL for better results\n")
	} else if output.AWXCache.HitRate >= 30 {
		builder.WriteString("  ⚠️  Moderate cache performance - review caching strategy\n")
	} else {
		builder.WriteString("  ❌ Low cache performance - cache warming may be needed\n")
	}

	totalRequests := output.AWXCache.Hits + output.AWXCache.Misses
	if totalRequests > 0 {
		savedRequests := output.AWXCache.Hits
		builder.WriteString(fmt.Sprintf("  • Backend requests saved: %d (%.1f%% reduction)\n",
			savedRequests, float64(savedRequests)/float64(totalRequests)*100))
	}

	builder.WriteString(fmt.Sprintf("\n📅 **Collected at:** %s\n", output.Timestamp))

	// Add full JSON response
	resultJSON, _ := json.MarshalIndent(output, "", "  ")
	builder.WriteString(fmt.Sprintf("\n**Full Response:**\n```json\n%s\n```", string(resultJSON)))

	return mcp.NewToolResultText(builder.String()), nil
}
