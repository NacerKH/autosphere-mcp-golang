package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NacerKH/autosphere-mcp-golang/internal/models"
	"github.com/NacerKH/autosphere-mcp-golang/internal/prometheus"
)

type ObservabilityService struct {
	prometheusClient *prometheus.PrometheusClient
}

func NewObservabilityService(prometheusURL, username, password string) *ObservabilityService {
	var promClient *prometheus.PrometheusClient
	
	if prometheusURL != "" {
		promClient = prometheus.NewPrometheusClient(prometheus.PrometheusConfig{
			BaseURL:  prometheusURL,
			Username: username,
			Password: password,
			Timeout:  30 * time.Second,
		})
	}

	return &ObservabilityService{
		prometheusClient: promClient,
	}
}

func (s *ObservabilityService) QueryPrometheus(ctx context.Context, args models.QueryPrometheusArgs) (models.QueryPrometheusOutput, error) {
	if s.prometheusClient == nil {
		return models.QueryPrometheusOutput{}, fmt.Errorf("Prometheus client not configured")
	}

	if args.Query == "" {
		return models.QueryPrometheusOutput{}, fmt.Errorf("query is required")
	}

	log.Printf("Executing Prometheus query: %s", args.Query)

	var result *prometheus.QueryResponse
	var err error

	// Determine if this is a range query
	if args.StartTime != "" && args.EndTime != "" {
		start, err := time.Parse(time.RFC3339, args.StartTime)
		if err != nil {
			return models.QueryPrometheusOutput{}, fmt.Errorf("invalid start_time format: %w", err)
		}

		end, err := time.Parse(time.RFC3339, args.EndTime)
		if err != nil {
			return models.QueryPrometheusOutput{}, fmt.Errorf("invalid end_time format: %w", err)
		}

		step, err := time.ParseDuration(args.Step)
		if err != nil {
			step = 5 * time.Minute // Default step
		}

		result, err = s.prometheusClient.QueryRange(ctx, args.Query, start, end, step)
	} else {
		result, err = s.prometheusClient.Query(ctx, args.Query)
	}

	if err != nil {
		log.Printf("Prometheus query failed: %v", err)
		return models.QueryPrometheusOutput{}, fmt.Errorf("query failed: %w", err)
	}

	// Convert results to our format
	metrics := make([]models.PrometheusMetric, 0, len(result.Data.Result))
	for _, res := range result.Data.Result {
		metric := models.PrometheusMetric{
			Labels: res.Metric,
		}

		if len(res.Value) >= 2 {
			metric.Value = res.Value[1].(string)
			if timestamp, ok := res.Value[0].(float64); ok {
				metric.Time = time.Unix(int64(timestamp), 0).Format(time.RFC3339)
			}
		}

		metrics = append(metrics, metric)
	}

	// Generate summary
	summary := s.generateQuerySummary(args.Query, result.Data.ResultType, len(metrics))

	log.Printf("Prometheus query returned %d metrics", len(metrics))

	return models.QueryPrometheusOutput{
		Query:      args.Query,
		ResultType: result.Data.ResultType,
		Metrics:    metrics,
		Summary:    summary,
		QueryTime:  time.Now().Format(time.RFC3339),
	}, nil
}

func (s *ObservabilityService) GetSystemMetrics(ctx context.Context, args models.GetSystemMetricsArgs) (models.GetSystemMetricsOutput, error) {
	if s.prometheusClient == nil {
		return models.GetSystemMetricsOutput{}, fmt.Errorf("Prometheus client not configured")
	}

	log.Printf("Retrieving system metrics")

	metrics, err := s.prometheusClient.GetSystemMetrics(ctx)
	if err != nil {
		log.Printf("Failed to get system metrics: %v", err)
		return models.GetSystemMetricsOutput{}, fmt.Errorf("failed to get system metrics: %w", err)
	}

	// Determine overall health
	overallHealth := "healthy"
	var alerts []string
	var recommendations []string

	// Check critical thresholds
	if cpu, ok := metrics["cpu_usage_percent"]; ok && cpu > 90 {
		overallHealth = "critical"
		alerts = append(alerts, fmt.Sprintf("High CPU usage: %.1f%%", cpu))
		recommendations = append(recommendations, "Consider scaling up or optimizing CPU-intensive processes")
	} else if cpu > 80 {
		overallHealth = "warning"
		alerts = append(alerts, fmt.Sprintf("Elevated CPU usage: %.1f%%", cpu))
	}

	if memory, ok := metrics["memory_usage_percent"]; ok && memory > 95 {
		overallHealth = "critical"
		alerts = append(alerts, fmt.Sprintf("High memory usage: %.1f%%", memory))
		recommendations = append(recommendations, "Consider adding memory or optimizing memory usage")
	} else if memory > 85 {
		if overallHealth == "healthy" {
			overallHealth = "warning"
		}
		alerts = append(alerts, fmt.Sprintf("Elevated memory usage: %.1f%%", memory))
	}

	if disk, ok := metrics["disk_usage_percent"]; ok && disk > 95 {
		overallHealth = "critical"
		alerts = append(alerts, fmt.Sprintf("High disk usage: %.1f%%", disk))
		recommendations = append(recommendations, "Clean up disk space or add storage")
	} else if disk > 85 {
		if overallHealth == "healthy" {
			overallHealth = "warning"
		}
		alerts = append(alerts, fmt.Sprintf("Elevated disk usage: %.1f%%", disk))
	}

	log.Printf("System health: %s, %d alerts", overallHealth, len(alerts))

	return models.GetSystemMetricsOutput{
		OverallHealth:   overallHealth,
		Metrics:         metrics,
		Alerts:          alerts,
		Timestamp:       time.Now().Format(time.RFC3339),
		Recommendations: recommendations,
	}, nil
}

func (s *ObservabilityService) GetAlerts(ctx context.Context, args models.GetAlertsArgs) (models.GetAlertsOutput, error) {
	if s.prometheusClient == nil {
		// Return mock alerts if Prometheus is not configured
		return s.getMockAlerts(), nil
	}

	log.Printf("Retrieving alerts from Prometheus")

	// For now, we'll create mock alerts since AlertManager integration is more complex
	// In a real implementation, you'd query AlertManager API
	return s.getMockAlerts(), nil
}

func (s *ObservabilityService) getMockAlerts() models.GetAlertsOutput {
	alerts := []models.AlertSummary{
		{
			Name:     "HighCPUUsage",
			Severity: "warning",
			Status:   "firing",
			Labels: map[string]string{
				"instance": "node-1",
				"job":      "node-exporter",
			},
			Annotations: map[string]string{
				"description": "CPU usage is above 80% for more than 5 minutes",
				"summary":     "High CPU usage detected",
			},
			ActiveSince: "15m",
			Value:       "85.2%",
		},
		{
			Name:     "DiskSpaceLow",
			Severity: "critical",
			Status:   "firing",
			Labels: map[string]string{
				"instance":   "server-prod-1",
				"mountpoint": "/",
			},
			Annotations: map[string]string{
				"description": "Disk space is running low",
				"summary":     "Less than 10% disk space remaining",
			},
			ActiveSince: "2h",
			Value:       "92.1%",
		},
	}

	critical := 0
	warning := 0
	for _, alert := range alerts {
		switch alert.Severity {
		case "critical":
			critical++
		case "warning":
			warning++
		}
	}

	summary := fmt.Sprintf("%d total alerts (%d critical, %d warning)", len(alerts), critical, warning)

	return models.GetAlertsOutput{
		ActiveAlerts: alerts,
		TotalAlerts:  len(alerts),
		Critical:     critical,
		Warning:      warning,
		Summary:      summary,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
}

func (s *ObservabilityService) generateQuerySummary(query, resultType string, count int) string {
	if count == 0 {
		return "Query returned no results"
	}

	// Detect common query patterns and provide meaningful summaries
	queryLower := strings.ToLower(query)

	if strings.Contains(queryLower, "cpu") {
		return fmt.Sprintf("Found %d CPU metrics from %s query", count, resultType)
	}
	if strings.Contains(queryLower, "memory") || strings.Contains(queryLower, "mem") {
		return fmt.Sprintf("Found %d memory metrics from %s query", count, resultType)
	}
	if strings.Contains(queryLower, "disk") || strings.Contains(queryLower, "filesystem") {
		return fmt.Sprintf("Found %d disk/filesystem metrics from %s query", count, resultType)
	}
	if strings.Contains(queryLower, "network") || strings.Contains(queryLower, "net") {
		return fmt.Sprintf("Found %d network metrics from %s query", count, resultType)
	}
	if strings.Contains(queryLower, "http") || strings.Contains(queryLower, "request") {
		return fmt.Sprintf("Found %d HTTP/request metrics from %s query", count, resultType)
	}

	return fmt.Sprintf("Query returned %d %s results", count, resultType)
}
