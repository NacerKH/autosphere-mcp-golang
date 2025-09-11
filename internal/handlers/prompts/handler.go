package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

type PromptsHandler struct {
	// Add any dependencies you need
}

func NewPromptsHandler() *PromptsHandler {
	return &PromptsHandler{}
}

// DeploymentPlanning provides structured guidance for deployment planning
func (h *PromptsHandler) DeploymentPlanning(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Extract arguments with default values
	environment := "production" // default
	version := "latest"         // default
	components := "all"         // default

	if request.Params.Arguments != nil {
		if env := request.Params.Arguments["environment"]; env != "" {
			environment = env
		}
		if ver := request.Params.Arguments["version"]; ver != "" {
			version = ver
		}
		if comp := request.Params.Arguments["components"]; comp != "" {
			components = comp
		}
	}

	promptText := fmt.Sprintf(`# Deployment Planning Guide for Autosphere

## Deployment Details
- **Environment**: %s
- **Version**: %s  
- **Components**: %s

## Pre-Deployment Checklist
1. **Infrastructure Readiness**
   - [ ] Kubernetes cluster is healthy and accessible
   - [ ] Required namespaces exist
   - [ ] Container registry is accessible
   - [ ] Database migrations are ready
   - [ ] External dependencies are available

2. **Security & Access**
   - [ ] Service accounts have proper permissions
   - [ ] Secrets and ConfigMaps are updated
   - [ ] Network policies are configured
   - [ ] TLS certificates are valid

3. **Monitoring & Observability**
   - [ ] Monitoring systems are operational
   - [ ] Log aggregation is working
   - [ ] Alerting rules are configured
   - [ ] Dashboards are accessible

## Recommended AWX Job Templates
Based on your deployment requirements, consider using these AWX job templates:

1. **autosphere-deploy** - Main deployment job
   - Use for: Deploying application components
   - Required variables: namespace, image_tag, replicas

2. **autosphere-health-check** - Post-deployment validation
   - Use for: Verifying deployment success
   - Required variables: check_endpoints, check_databases

## Deployment Strategy for %s
- **Rolling Update**: Recommended for production to ensure zero-downtime
- **Blue-Green**: Consider for major version upgrades
- **Canary**: Use when introducing significant changes

## Post-Deployment Validation
1. Health check all services
2. Verify database connections
3. Test critical user flows
4. Monitor performance metrics
5. Check log outputs for errors

## Rollback Plan
If issues are detected:
1. Stop deployment immediately
2. Revert to previous version using AWX
3. Verify system stability
4. Investigate root cause

Would you like me to help you execute any of these steps using the available AWX tools?`, 
		environment, version, components, environment)

	return mcp.NewGetPromptResult(
		fmt.Sprintf("Deployment planning for %s environment", environment),
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(promptText),
			),
		},
	), nil
}

// TroubleshootingGuide provides systematic troubleshooting guidance
func (h *PromptsHandler) TroubleshootingGuide(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	issue := "Unknown issue"
	component := "system"

	if request.Params.Arguments != nil {
		if iss := request.Params.Arguments["issue"]; iss != "" {
			issue = iss
		}
		if comp := request.Params.Arguments["component"]; comp != "" {
			component = comp
		}
	}

	promptText := fmt.Sprintf(`# Troubleshooting Guide: %s

## Problem Summary
- **Issue**: %s
- **Component**: %s
- **Timestamp**: $(current_time)

## Initial Assessment
Let's start with basic diagnostics for the %s component:

### Step 1: Check Component Health
Run the health check tool to get current status:
- Use the **health_check** tool to verify overall system health
- Focus on the %s component specifically

### Step 2: Review Recent Changes
Consider recent deployments or configuration changes:
- Check AWX job history using **list_awx_jobs**
- Look for failed or recent jobs that might have caused issues

### Step 3: Examine Logs and Metrics
Investigate system logs and performance metrics:
- Check application logs for error patterns
- Review resource utilization (CPU, memory, network)
- Verify database connectivity and performance

## Common Causes for %s Issues
Based on the component type, here are typical causes:

### Application Components (API, Web, Workers)
- Memory leaks or high resource usage
- Database connection pool exhaustion
- External service dependencies failing
- Configuration errors after deployment

### Infrastructure Components (Database, Cache)
- Disk space issues
- Network connectivity problems
- Resource limits exceeded
- Backup or maintenance operations

### Monitoring Components
- Metrics collection failures
- Alert rule misconfigurations
- Dashboard connectivity issues

## Systematic Diagnosis Steps
1. **Immediate Response**
   - Check if this is a widespread issue affecting multiple components
   - Verify if automatic scaling has been triggered
   - Look for active alerts in monitoring systems

2. **Deep Investigation**
   - Use **get_job_output** to review recent AWX job logs
   - Check **list_awx_resources** to verify infrastructure state
   - Examine component-specific metrics and logs

3. **Resolution Planning**
   - If scaling issue: Use **autoscale** tool to adjust resources
   - If deployment issue: Consider rollback using AWX deployment jobs
   - If configuration issue: Review and update configurations

## Recovery Actions
Based on findings, choose appropriate recovery method:
- **Scaling**: Increase resources if performance-related
- **Restart**: Restart affected services if temporary issue
- **Rollback**: Revert to last known good state if deployment-related
- **Hotfix**: Apply targeted fix if specific bug identified

Would you like me to help you execute any of these diagnostic steps using the available tools?`,
		issue, issue, component, component, component, component)

	return mcp.NewGetPromptResult(
		fmt.Sprintf("Troubleshooting guide for %s issue in %s", issue, component),
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(promptText),
			),
		},
	), nil
}
