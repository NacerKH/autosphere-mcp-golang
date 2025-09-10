package awx

import "time"

// AWX API Response Models

type AuthResponse struct {
	Token string `json:"token"`
}

type JobTemplate struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Inventory   int    `json:"inventory"`
	Project     int    `json:"project"`
	Playbook    string `json:"playbook"`
}

type JobTemplateList struct {
	Count   int           `json:"count"`
	Results []JobTemplate `json:"results"`
}

type LaunchJobRequest struct {
	ExtraVars map[string]interface{} `json:"extra_vars,omitempty"`
	Inventory int                    `json:"inventory,omitempty"`
	Limit     string                 `json:"limit,omitempty"`
	Tags      string                 `json:"job_tags,omitempty"`
	SkipTags  string                 `json:"skip_tags,omitempty"`
}

type Job struct {
	ID              int                    `json:"id"`
	Name            string                 `json:"name"`
	Status          string                 `json:"status"`
	Started         *time.Time             `json:"started"`
	Finished        *time.Time             `json:"finished"`
	Elapsed         float64                `json:"elapsed"`
	JobTemplate     int                    `json:"job_template"`
	PlaybookResults map[string]interface{} `json:"job_events,omitempty"`
	URL             string                 `json:"url"`
}

type JobLaunchResponse struct {
	Job                int    `json:"job"`
	IgnoredFields      map[string]interface{} `json:"ignored_fields"`
	ID                 int    `json:"id"`
	JobTemplate        int    `json:"job_template"`
	URL                string `json:"url"`
	RelatedJobTemplate string `json:"related_job_template"`
}

type ErrorResponse struct {
	Detail string `json:"detail"`
}
