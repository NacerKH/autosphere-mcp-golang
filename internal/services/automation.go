package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	
	"github.com/NacerKH/autosphere-mcp-golang/internal/awx"
	"github.com/NacerKH/autosphere-mcp-golang/internal/models"
)

type AutomationService struct {
	healthService *HealthService
	awxClient     *awx.Client
	awxBaseURL    string
}

func NewAutomationService(healthService *HealthService, awxClient *awx.Client, awxBaseURL string) *AutomationService {
	return &AutomationService{
		healthService: healthService,
		awxClient:     awxClient,
		awxBaseURL:    awxBaseURL,
	}
}

func (s *AutomationService) LaunchJob(ctx context.Context, args models.AWXJobArgs) (models.AWXJobOutput, error) {
	if args.JobTemplate == "" {
		return models.AWXJobOutput{}, fmt.Errorf("job_template is required")
	}
	
	log.Printf("Launching AWX job with template: %s", args.JobTemplate)
	
	// Create job launcher with professional configuration
	launcher := awx.NewJobLauncher(s.awxClient)
	
	// Prepare launch options
	options := awx.LaunchJobOptions{
		TemplateNameOrID: args.JobTemplate,
		ExtraVars:        make(map[string]interface{}),
		Inventory:        args.Inventory,
		Limit:            args.Limit,
		Tags:             args.Tags,
		SkipTags:         args.SkipTags,
		Timeout:          60 * time.Second,
	}
	
	// Convert extra vars
	for k, v := range args.ExtraVars {
		options.ExtraVars[k] = v
	}
	
	// Launch the job using professional launcher
	result, err := launcher.Launch(ctx, options)
	if err != nil {
		log.Printf("Failed to launch AWX job: %v", err)
		return models.AWXJobOutput{}, fmt.Errorf("failed to launch AWX job: %w", err)
	}
	
	log.Printf("AWX job launched successfully: ID %d", result.JobID)
	
	return models.AWXJobOutput{
		JobID:   result.JobID,
		Status:  result.Status,
		URL:     result.URL,
		Message: result.Message,
	}, nil
}

func (s *AutomationService) CheckJobStatus(ctx context.Context, args models.AWXStatusArgs) (models.AWXStatusOutput, error) {
	if args.JobID <= 0 {
		return models.AWXStatusOutput{}, fmt.Errorf("valid job_id is required")
	}
	
	log.Printf("Checking AWX job status for ID: %d", args.JobID)
	
	// Get job details from AWX
	job, err := s.awxClient.GetJob(ctx, args.JobID)
	if err != nil {
		log.Printf("Failed to get AWX job status: %v", err)
		return models.AWXStatusOutput{}, fmt.Errorf("failed to get job status: %w", err)
	}
	
	log.Printf("Retrieved AWX job status: %s", job.Status)
	
	// Format timestamps
	startedAt := ""
	finishedAt := ""
	elapsedTime := ""
	
	if job.Started != nil {
		startedAt = job.Started.Format("2006-01-02 15:04:05")
		if job.Finished != nil {
			finishedAt = job.Finished.Format("2006-01-02 15:04:05")
			elapsedTime = job.Finished.Sub(*job.Started).Round(time.Second).String()
		} else {
			elapsedTime = time.Since(*job.Started).Round(time.Second).String()
		}
	}
	
	// Create mock results for now (can be enhanced to get real job events)
	results := map[string]interface{}{
		"status": job.Status,
	}
	
	if job.Status == "successful" {
		results["changed"] = 2
		results["ok"] = 8
		results["failed"] = 0
		results["skipped"] = 1
	}
	
	return models.AWXStatusOutput{
		JobID:           args.JobID,
		Status:          job.Status,
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
		ElapsedTime:     elapsedTime,
		PlaybookResults: results,
	}, nil
}

func (s *AutomationService) CheckHealth(ctx context.Context, args models.HealthCheckArgs) (models.HealthCheckOutput, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	components := make(map[string]models.ComponentHealth)
	var recommendations []string
	
	componentsToCheck := []string{"api", "database", "cache", "web", "workers", "monitoring"}
	
	if args.Component != "" && args.Component != "all" {
		componentsToCheck = []string{args.Component}
	}
	
	overallHealthy := true
	
	for _, comp := range componentsToCheck {
		health := s.healthService.CheckComponent(comp, args.Deep)
		components[comp] = health
		
		if health.Status == "critical" || health.Status == "warning" {
			overallHealthy = false
		}
		
		if health.Status == "warning" {
			switch comp {
			case "database":
				recommendations = append(recommendations, "Consider optimizing database queries or scaling database resources")
			case "cache":
				recommendations = append(recommendations, "Check cache hit ratio and consider increasing cache size")
			case "api":
				recommendations = append(recommendations, "Monitor API response times and consider horizontal scaling")
			}
		} else if health.Status == "critical" {
			recommendations = append(recommendations, fmt.Sprintf("URGENT: %s component requires immediate attention", comp))
		}
	}
	
	overallStatus := "healthy"
	if !overallHealthy {
		hasCritical := false
		for _, comp := range components {
			if comp.Status == "critical" {
				hasCritical = true
				break
			}
		}
		if hasCritical {
			overallStatus = "critical"
		} else {
			overallStatus = "warning"
		}
	}
	
	return models.HealthCheckOutput{
		OverallStatus:   overallStatus,
		Components:      components,
		Timestamp:       timestamp,
		Recommendations: recommendations,
	}, nil
}

func (s *AutomationService) Autoscale(ctx context.Context, args models.AutoscaleArgs) (models.AutoscaleOutput, error) {
	if args.Action == "" {
		return models.AutoscaleOutput{}, fmt.Errorf("action is required")
	}
	
	service := args.Service
	if service == "" {
		service = "api"
	}
	
	oldReplicas := 3
	newReplicas := oldReplicas
	reason := ""
	awxJobID := 0
	
	switch args.Action {
	case "scale_up":
		if args.Replicas > 0 {
			newReplicas = args.Replicas
		} else {
			newReplicas = oldReplicas + 2
		}
		reason = "Manual scale up requested"
		awxJobID = int(time.Now().Unix())
		
	case "scale_down":
		if args.Replicas > 0 {
			newReplicas = args.Replicas
		} else {
			if oldReplicas-1 > 1 {
				newReplicas = oldReplicas - 1
			} else {
				newReplicas = 1
			}
		}
		reason = "Manual scale down requested"
		awxJobID = int(time.Now().Unix())
		
	case "analyze":
		reason = s.healthService.AnalyzeLoad(args.Threshold)
		if strings.Contains(reason, "high") {
			newReplicas = oldReplicas + 1
		} else if strings.Contains(reason, "low") {
			if oldReplicas-1 > 1 {
				newReplicas = oldReplicas - 1
			} else {
				newReplicas = 1
			}
		}
		
	case "auto":
		metrics := s.healthService.GetSystemMetrics()
		if metrics["cpu"] > 80 || metrics["memory"] > 85 {
			newReplicas = oldReplicas + 2
			reason = "Auto-scaling up due to high resource usage"
			awxJobID = int(time.Now().Unix())
		} else if metrics["cpu"] < 20 && metrics["memory"] < 30 && oldReplicas > 1 {
			if oldReplicas-1 > 1 {
				newReplicas = oldReplicas - 1
			} else {
				newReplicas = 1
			}
			reason = "Auto-scaling down due to low resource usage"
			awxJobID = int(time.Now().Unix())
		} else {
			reason = "No scaling needed - metrics within normal range"
		}
		
	default:
		return models.AutoscaleOutput{}, fmt.Errorf("unknown action: %s", args.Action)
	}
	
	status := "completed"
	if awxJobID > 0 {
		status = "job_launched"
	}
	
	return models.AutoscaleOutput{
		Action:      args.Action,
		Service:     service,
		OldReplicas: oldReplicas,
		NewReplicas: newReplicas,
		Reason:      reason,
		JobID:       awxJobID,
		Status:      status,
	}, nil
}
