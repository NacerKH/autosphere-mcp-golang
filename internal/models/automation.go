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
