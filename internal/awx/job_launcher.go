package awx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type JobLauncher struct {
	client *Client
}

func NewJobLauncher(client *Client) *JobLauncher {
	return &JobLauncher{client: client}
}

type LaunchJobOptions struct {
	TemplateNameOrID string
	ExtraVars        map[string]interface{}
	Inventory        string
	Limit            string
	Tags             string
	SkipTags         string
	JobType          string
	Verbosity        int
	DiffMode         bool
	Timeout          time.Duration
}

type LaunchResult struct {
	JobID      int    `json:"job_id"`
	Status     string `json:"status"`
	URL        string `json:"url"`
	Message    string `json:"message"`
	LaunchType string `json:"launch_type"`
}

func (jl *JobLauncher) Launch(ctx context.Context, options LaunchJobOptions) (*LaunchResult, error) {
	if options.TemplateNameOrID == "" {
		return nil, fmt.Errorf("template name or ID is required")
	}

	if options.Timeout == 0 {
		options.Timeout = 60 * time.Second
	}

	log.Printf("Starting job launch for template: %s", options.TemplateNameOrID)

	templateID, templateName, err := jl.resolveTemplate(ctx, options.TemplateNameOrID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve template '%s': %w", options.TemplateNameOrID, err)
	}

	log.Printf("Resolved template '%s' to ID: %d", templateName, templateID)

	if err := jl.validateLaunchPermissions(ctx, templateID); err != nil {
		return nil, fmt.Errorf("launch validation failed for template %d: %w", templateID, err)
	}

	launchRequest := jl.prepareLaunchRequest(options)

	response, err := jl.executeLaunchWithRetry(ctx, templateID, launchRequest, options.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to launch job for template %d: %w", templateID, err)
	}

	result := &LaunchResult{
		JobID:      response.Job,
		Status:     "pending",
		URL:        response.URL,
		LaunchType: "api",
		Message:    jl.createSuccessMessage(templateName, response.Job, options),
	}

	log.Printf("Job launched successfully: ID %d, Template: %s", result.JobID, templateName)
	return result, nil
}

func (jl *JobLauncher) resolveTemplate(ctx context.Context, nameOrID string) (int, string, error) {
	templates, err := jl.client.GetJobTemplates(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch job templates: %w", err)
	}

	if id, err := strconv.Atoi(nameOrID); err == nil {
		for _, template := range templates {
			if template.ID == id {
				return template.ID, template.Name, nil
			}
		}
	}

	for _, template := range templates {
		if template.Name == nameOrID {
			return template.ID, template.Name, nil
		}
	}

	var available []string
	for _, template := range templates {
		available = append(available, fmt.Sprintf("'%s' (ID: %d)", template.Name, template.ID))
	}

	return 0, "", fmt.Errorf("template not found. Available templates: %s", strings.Join(available, ", "))
}

func (jl *JobLauncher) validateLaunchPermissions(ctx context.Context, templateID int) error {
	url := fmt.Sprintf("%s/api/v2/job_templates/%d/launch/", jl.client.baseURL, templateID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	if jl.client.token != "" {
		req.Header.Set("Authorization", "Bearer "+jl.client.token)
	} else if jl.client.username != "" && jl.client.password != "" {
		req.SetBasicAuth(jl.client.username, jl.client.password)
	}
	
	resp, err := jl.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate launch permissions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return fmt.Errorf("insufficient permissions to launch this job template")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("validation failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (jl *JobLauncher) prepareLaunchRequest(options LaunchJobOptions) map[string]interface{} {
	request := make(map[string]interface{})

	if len(options.ExtraVars) > 0 {
		request["extra_vars"] = options.ExtraVars
	}
	if options.Inventory != "" {
		request["inventory"] = options.Inventory
	}
	if options.Limit != "" {
		request["limit"] = options.Limit
	}
	if options.Tags != "" {
		request["job_tags"] = options.Tags
	}
	if options.SkipTags != "" {
		request["skip_tags"] = options.SkipTags
	}
	if options.JobType != "" {
		request["job_type"] = options.JobType
	}
	if options.Verbosity > 0 {
		request["verbosity"] = options.Verbosity
	}
	if options.DiffMode {
		request["diff_mode"] = true
	}

	return request
}

func (jl *JobLauncher) executeLaunchWithRetry(ctx context.Context, templateID int, request map[string]interface{}, timeout time.Duration) (*JobLaunchResponse, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Launch attempt %d/%d for template %d", attempt, maxRetries, templateID)
		
		response, err := jl.executeSingleLaunch(ctx, templateID, request, timeout)
		if err == nil {
			return response, nil
		}

		lastErr = err
		log.Printf("Launch attempt %d failed: %v", attempt, err)

		if isNonRetryableError(err) {
			break
		}

		if attempt < maxRetries {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("all launch attempts failed, last error: %w", lastErr)
}

func (jl *JobLauncher) executeSingleLaunch(ctx context.Context, templateID int, request map[string]interface{}, timeout time.Duration) (*JobLaunchResponse, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	url := fmt.Sprintf("%s/api/v2/job_templates/%d/launch/", jl.client.baseURL, templateID)
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("POST %s", url)
	log.Printf("Request body: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if jl.client.token != "" {
		req.Header.Set("Authorization", "Bearer "+jl.client.token)
	} else if jl.client.username != "" && jl.client.password != "" {
		req.SetBasicAuth(jl.client.username, jl.client.password)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("Response: %d - %s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("AWX API error: status %d - %s", resp.StatusCode, string(respBody))
	}

	var response JobLaunchResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (jl *JobLauncher) createSuccessMessage(templateName string, jobID int, options LaunchJobOptions) string {
	message := fmt.Sprintf("Successfully launched job %d using template '%s'", jobID, templateName)
	
	if len(options.ExtraVars) > 0 {
		message += fmt.Sprintf(" with %d extra variables", len(options.ExtraVars))
	}
	
	if options.Limit != "" {
		message += fmt.Sprintf(" limited to hosts: %s", options.Limit)
	}
	
	return message
}

func isNonRetryableError(err error) bool {
	errStr := err.Error()
	
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "insufficient permissions") {
		return true
	}
	
	if strings.Contains(errStr, "400") || strings.Contains(errStr, "not found") {
		return true
	}
	
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "authentication") {
		return true
	}
	
	return false
}
