package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	
	"github.com/NacerKH/autosphere-mcp-golang/internal/models"
)

type AutomationService struct {
	healthService *HealthService
	awxBaseURL    string
}

func NewAutomationService(healthService *HealthService, awxBaseURL string) *AutomationService {
	return &AutomationService{
		healthService: healthService,
		awxBaseURL:    awxBaseURL,
	}
}

func (s *AutomationService) LaunchJob(ctx context.Context, args models.AWXJobArgs) (models.AWXJobOutput, error) {
	if args.JobTemplate == "" {
		return models.AWXJobOutput{}, fmt.Errorf("job_template is required")
	}
	
	jobID := int(time.Now().Unix())
	
	extraVarsStr := ""
	if len(args.ExtraVars) > 0 {
		extraVarsBytes, _ := json.Marshal(args.ExtraVars)
		extraVarsStr = string(extraVarsBytes)
	}
	
	var message string
	switch strings.ToLower(args.JobTemplate) {
	case "autosphere-deploy", "deploy-autosphere":
		message = fmt.Sprintf("Launched Autosphere deployment job (ID: %d) with template '%s'", jobID, args.JobTemplate)
	case "autosphere-health-check", "health-check":
		message = fmt.Sprintf("Launched Autosphere health check job (ID: %d)", jobID)
	case "autosphere-scale", "scale-services":
		message = fmt.Sprintf("Launched Autosphere scaling job (ID: %d) with vars: %s", jobID, extraVarsStr)
	case "autosphere-backup", "backup":
		message = fmt.Sprintf("Launched Autosphere backup job (ID: %d)", jobID)
	case "autosphere-update", "update":
		message = fmt.Sprintf("Launched Autosphere update job (ID: %d)", jobID)
	default:
		message = fmt.Sprintf("Launched job (ID: %d) with template '%s'", jobID, args.JobTemplate)
	}
	
	return models.AWXJobOutput{
		JobID:   jobID,
		Status:  "pending",
		URL:     fmt.Sprintf("%s/#/jobs/playbook/%d", s.awxBaseURL, jobID),
		Message: message,
	}, nil
}

func (s *AutomationService) CheckJobStatus(ctx context.Context, args models.AWXStatusArgs) (models.AWXStatusOutput, error) {
	if args.JobID <= 0 {
		return models.AWXStatusOutput{}, fmt.Errorf("valid job_id is required")
	}
	
	now := time.Now()
	startTime := time.Unix(int64(args.JobID), 0)
	elapsed := now.Sub(startTime)
	
	var status string
	var finishedAt string
	var results map[string]interface{}
	
	if elapsed < 30*time.Second {
		status = "running"
	} else if elapsed < 60*time.Second {
		status = "successful"
		finishedAt = startTime.Add(30 * time.Second).Format("2006-01-02 15:04:05")
		results = map[string]interface{}{
			"changed": 3,
			"ok":      15,
			"failed":  0,
			"skipped": 2,
		}
	} else {
		status = "successful"
		finishedAt = startTime.Add(45 * time.Second).Format("2006-01-02 15:04:05")
		results = map[string]interface{}{
			"changed": 5,
			"ok":      20,
			"failed":  0,
			"skipped": 1,
		}
	}
	
	return models.AWXStatusOutput{
		JobID:           args.JobID,
		Status:          status,
		StartedAt:       startTime.Format("2006-01-02 15:04:05"),
		FinishedAt:      finishedAt,
		ElapsedTime:     elapsed.Round(time.Second).String(),
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
