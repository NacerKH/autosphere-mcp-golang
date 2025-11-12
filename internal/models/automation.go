package models

type AWXJobArgs struct {
	JobTemplate   string            `json:"job_template" jsonschema:"the name or ID of the AWX job template"`
	ExtraVars     map[string]string `json:"extra_vars,omitempty" jsonschema:"extra variables to pass to the job"`
	Inventory     string            `json:"inventory,omitempty" jsonschema:"inventory name or ID (optional)"`
	Limit         string            `json:"limit,omitempty" jsonschema:"limit the job to specific hosts (optional)"`
	Tags          string            `json:"tags,omitempty" jsonschema:"ansible tags to run (optional)"`
	SkipTags      string            `json:"skip_tags,omitempty" jsonschema:"ansible tags to skip (optional)"`
}

type AWXJobOutput struct {
	JobID     int    `json:"job_id" jsonschema:"the AWX job ID"`
	Status    string `json:"status" jsonschema:"the job status"`
	URL       string `json:"url" jsonschema:"the AWX job URL"`
	Message   string `json:"message" jsonschema:"human-readable status message"`
}

type AWXStatusArgs struct {
	JobID int `json:"job_id" jsonschema:"the AWX job ID to check"`
}

type AWXStatusOutput struct {
	JobID           int                    `json:"job_id" jsonschema:"the AWX job ID"`
	Status          string                 `json:"status" jsonschema:"the current job status"`
	StartedAt       string                 `json:"started_at" jsonschema:"when the job started"`
	FinishedAt      string                 `json:"finished_at,omitempty" jsonschema:"when the job finished (if completed)"`
	ElapsedTime     string                 `json:"elapsed_time" jsonschema:"how long the job has been running"`
	PlaybookResults map[string]interface{} `json:"playbook_results,omitempty" jsonschema:"results from the playbook execution"`
}

type HealthCheckArgs struct {
	Component string `json:"component,omitempty" jsonschema:"specific component to check (api, database, cache, all)"`
	Deep      bool   `json:"deep,omitempty" jsonschema:"perform deep health checks (default: false)"`
}

type HealthCheckOutput struct {
	OverallStatus   string                      `json:"overall_status" jsonschema:"overall system health status"`
	Components      map[string]ComponentHealth  `json:"components" jsonschema:"health status of individual components"`
	Timestamp       string                      `json:"timestamp" jsonschema:"when the health check was performed"`
	Recommendations []string                    `json:"recommendations,omitempty" jsonschema:"suggested actions based on health status"`
}

type ComponentHealth struct {
	Status      string            `json:"status" jsonschema:"component status (healthy, warning, critical, unknown)"`
	Details     string            `json:"details" jsonschema:"detailed status information"`
	Metrics     map[string]string `json:"metrics,omitempty" jsonschema:"relevant metrics for this component"`
	LastChecked string            `json:"last_checked" jsonschema:"when this component was last checked"`
}

type AutoscaleArgs struct {
	Action    string `json:"action" jsonschema:"autoscaling action (scale_up, scale_down, analyze, auto)"`
	Service   string `json:"service,omitempty" jsonschema:"specific service to scale (optional)"`
	Replicas  int    `json:"replicas,omitempty" jsonschema:"target number of replicas (for manual scaling)"`
	Threshold string `json:"threshold,omitempty" jsonschema:"scaling threshold (cpu_high, memory_high, load_high)"`
}

type AutoscaleOutput struct {
	Action      string `json:"action" jsonschema:"action taken"`
	Service     string `json:"service" jsonschema:"service affected"`
	OldReplicas int    `json:"old_replicas" jsonschema:"previous number of replicas"`
	NewReplicas int    `json:"new_replicas" jsonschema:"new number of replicas"`
	Reason      string `json:"reason" jsonschema:"reason for scaling decision"`
	JobID       int    `json:"job_id,omitempty" jsonschema:"AWX job ID if automation was triggered"`
	Status      string `json:"status" jsonschema:"operation status"`
}

// New models for additional tools

type ListJobsArgs struct {
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum number of jobs to return (default: 20)"`
	Status string `json:"status,omitempty" jsonschema:"filter by job status (successful, failed, running, pending)"`
}

type ListJobsOutput struct {
	Jobs []JobSummary `json:"jobs" jsonschema:"list of jobs"`
	Total int         `json:"total" jsonschema:"total number of jobs"`
}

type JobSummary struct {
	ID          int    `json:"id" jsonschema:"job ID"`
	Name        string `json:"name" jsonschema:"job name"`
	Status      string `json:"status" jsonschema:"job status"`
	Template    string `json:"template" jsonschema:"job template name"`
	StartedAt   string `json:"started_at,omitempty" jsonschema:"when the job started"`
	FinishedAt  string `json:"finished_at,omitempty" jsonschema:"when the job finished"`
	ElapsedTime string `json:"elapsed_time,omitempty" jsonschema:"job duration"`
}

type GetJobOutputArgs struct {
	JobID int `json:"job_id" jsonschema:"the AWX job ID to get output for"`
}

type GetJobOutputOutput struct {
	JobID  int    `json:"job_id" jsonschema:"the job ID"`
	Output string `json:"output" jsonschema:"the job output/logs"`
}

type CancelJobArgs struct {
	JobID int `json:"job_id" jsonschema:"the AWX job ID to cancel"`
}

type CancelJobOutput struct {
	JobID   int    `json:"job_id" jsonschema:"the job ID"`
	Status  string `json:"status" jsonschema:"cancellation status"`
	Message string `json:"message" jsonschema:"status message"`
}

type ListResourcesArgs struct {
	ResourceType string `json:"resource_type" jsonschema:"type of resource (templates, inventories, projects)"`
}

type ListResourcesOutput struct {
	ResourceType string        `json:"resource_type" jsonschema:"type of resource"`
	Resources    []interface{} `json:"resources" jsonschema:"list of resources"`
	Total        int           `json:"total" jsonschema:"total number of resources"`
}

type ResourceSummary struct {
	ID          int    `json:"id" jsonschema:"resource ID"`
	Name        string `json:"name" jsonschema:"resource name"`
	Description string `json:"description" jsonschema:"resource description"`
	Status      string `json:"status,omitempty" jsonschema:"resource status"`
}

// Prometheus and observability models

type QueryPrometheusArgs struct {
	Query     string `json:"query" jsonschema:"PromQL query to execute"`
	StartTime string `json:"start_time,omitempty" jsonschema:"start time for range queries (RFC3339 format)"`
	EndTime   string `json:"end_time,omitempty" jsonschema:"end time for range queries (RFC3339 format)"`
	Step      string `json:"step,omitempty" jsonschema:"step size for range queries (e.g., '5m', '1h')"`
}

type QueryPrometheusOutput struct {
	Query       string                 `json:"query" jsonschema:"the executed query"`
	ResultType  string                 `json:"result_type" jsonschema:"type of result (vector, matrix, scalar)"`
	Metrics     []PrometheusMetric     `json:"metrics" jsonschema:"list of metric results"`
	Summary     string                 `json:"summary" jsonschema:"human-readable summary of results"`
	QueryTime   string                 `json:"query_time" jsonschema:"when the query was executed"`
}

type PrometheusMetric struct {
	Labels map[string]string `json:"labels" jsonschema:"metric labels"`
	Value  string            `json:"value" jsonschema:"metric value"`
	Time   string            `json:"time" jsonschema:"timestamp of the value"`
}

type GetSystemMetricsArgs struct {
	TimeRange string `json:"time_range,omitempty" jsonschema:"time range for metrics (5m, 1h, 24h)"`
	Nodes     string `json:"nodes,omitempty" jsonschema:"specific nodes to query (comma-separated)"`
}

type GetSystemMetricsOutput struct {
	OverallHealth string                    `json:"overall_health" jsonschema:"overall system health status"`
	Metrics       map[string]float64        `json:"metrics" jsonschema:"key system metrics"`
	NodeMetrics   map[string]map[string]float64 `json:"node_metrics,omitempty" jsonschema:"per-node metrics"`
	Alerts        []string                  `json:"alerts,omitempty" jsonschema:"active alerts"`
	Timestamp     string                    `json:"timestamp" jsonschema:"when metrics were collected"`
	Recommendations []string                `json:"recommendations,omitempty" jsonschema:"optimization recommendations"`
}

type GetAlertsArgs struct {
	Severity string `json:"severity,omitempty" jsonschema:"filter by alert severity (critical, warning, info)"`
	Service  string `json:"service,omitempty" jsonschema:"filter by service name"`
	Active   *bool  `json:"active,omitempty" jsonschema:"filter by active status"`
}

type GetAlertsOutput struct {
	ActiveAlerts []AlertSummary `json:"active_alerts" jsonschema:"currently active alerts"`
	TotalAlerts  int           `json:"total_alerts" jsonschema:"total number of alerts"`
	Critical     int           `json:"critical_count" jsonschema:"number of critical alerts"`
	Warning      int           `json:"warning_count" jsonschema:"number of warning alerts"`
	Summary      string        `json:"summary" jsonschema:"overall alert status summary"`
	Timestamp    string        `json:"timestamp" jsonschema:"when alerts were retrieved"`
}

type AlertSummary struct {
	Name        string            `json:"name" jsonschema:"alert name"`
	Severity    string            `json:"severity" jsonschema:"alert severity"`
	Status      string            `json:"status" jsonschema:"alert status"`
	Labels      map[string]string `json:"labels" jsonschema:"alert labels"`
	Annotations map[string]string `json:"annotations" jsonschema:"alert annotations"`
	ActiveSince string            `json:"active_since,omitempty" jsonschema:"how long the alert has been active"`
	Value       string            `json:"value,omitempty" jsonschema:"current metric value"`
}

// Job Template models

type ListJobTemplatesArgs struct {
	// No arguments needed for simple list
}

type ListJobTemplatesOutput struct {
	Templates []JobTemplateSummary `json:"templates" jsonschema:"list of job templates"`
	Total     int                  `json:"total" jsonschema:"total number of templates"`
}

type JobTemplateSummary struct {
	ID          int    `json:"id" jsonschema:"template ID"`
	Name        string `json:"name" jsonschema:"template name"`
	Description string `json:"description" jsonschema:"template description"`
	Playbook    string `json:"playbook" jsonschema:"playbook path"`
	Inventory   int    `json:"inventory" jsonschema:"inventory ID"`
	Project     int    `json:"project" jsonschema:"project ID"`
}

type CreateJobTemplateArgs struct {
	Name        string `json:"name" jsonschema:"required,template name"`
	Description string `json:"description,omitempty" jsonschema:"template description"`
	JobType     string `json:"job_type,omitempty" jsonschema:"job type (default: run)"`
	Inventory   int    `json:"inventory" jsonschema:"required,inventory ID"`
	Project     int    `json:"project" jsonschema:"required,project ID"`
	Playbook    string `json:"playbook" jsonschema:"required,playbook path (e.g., site.yml)"`
	Verbosity   int    `json:"verbosity,omitempty" jsonschema:"playbook verbosity (0-5)"`
}

type CreateJobTemplateOutput struct {
	ID          int    `json:"id" jsonschema:"created template ID"`
	Name        string `json:"name" jsonschema:"template name"`
	Description string `json:"description" jsonschema:"template description"`
	Status      string `json:"status" jsonschema:"creation status"`
	Message     string `json:"message" jsonschema:"status message"`
}

// Cache statistics models

type GetCacheStatsArgs struct {
	// No arguments needed
}

type GetCacheStatsOutput struct {
	AWXCache        CacheStatsDetail `json:"awx_cache" jsonschema:"AWX client cache statistics"`
	PrometheusCache CacheStatsDetail `json:"prometheus_cache,omitempty" jsonschema:"Prometheus client cache statistics"`
	Summary         string           `json:"summary" jsonschema:"human-readable summary"`
	Timestamp       string           `json:"timestamp" jsonschema:"when statistics were collected"`
}

type CacheStatsDetail struct {
	Hits        int64   `json:"hits" jsonschema:"cache hits"`
	Misses      int64   `json:"misses" jsonschema:"cache misses"`
	Sets        int64   `json:"sets" jsonschema:"cache sets"`
	Evictions   int64   `json:"evictions" jsonschema:"cache evictions"`
	CurrentSize int     `json:"current_size" jsonschema:"current number of cached items"`
	HitRate     float64 `json:"hit_rate" jsonschema:"cache hit rate percentage"`
}
