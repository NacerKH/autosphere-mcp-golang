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

type Client struct {
	baseURL    string
	username   string
	password   string
	token      string
	httpClient *http.Client
}

type ClientConfig struct {
	BaseURL  string
	Username string
	Password string
	Token    string
	Timeout  time.Duration
}

func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL:  strings.TrimSuffix(config.BaseURL, "/"),
		username: config.Username,
		password: config.Password,
		token:    config.Token,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (c *Client) authenticate(ctx context.Context) error {
	if c.token != "" {
		log.Printf("Using provided AWX token")
		return nil
	}

	if c.username == "" || c.password == "" {
		return fmt.Errorf("either token or username/password must be provided")
	}

	log.Printf("Testing Basic Auth with AWX...")
	if err := c.testBasicAuth(ctx); err == nil {
		log.Printf("Basic Auth successful - using direct Basic Auth")
		return nil
	}

	log.Printf("Trying token creation via /api/v2/tokens/...")
	if err := c.createTokenViaTokensEndpoint(ctx); err == nil {
		return nil
	}

	log.Printf("Trying legacy /api/v2/authtoken/...")
	return c.createTokenViaLegacyEndpoint(ctx)
}

func (c *Client) testBasicAuth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/job_templates/", nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("basic auth test failed with status %d", resp.StatusCode)
}

func (c *Client) createTokenViaTokensEndpoint(ctx context.Context) error {
	tokenData := map[string]interface{}{
		"description": "MCP Autosphere Token",
		"scope":       "write",
	}

	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v2/tokens/", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	if token, ok := tokenResp["token"].(string); ok {
		c.token = token
		log.Printf("Successfully created AWX token")
		return nil
	}

	return fmt.Errorf("token not found in response")
}

func (c *Client) createTokenViaLegacyEndpoint(ctx context.Context) error {
	authData := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v2/authtoken/", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate with AWX: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.token = authResp.Token
	log.Printf("Successfully authenticated with AWX (legacy)")
	return nil
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	// Try to authenticate first if we don't have a token
	if c.token == "" && c.username != "" {
		// Just test basic auth, don't fail if it doesn't work
		c.testBasicAuth(ctx)
	}

	var reqBody io.Reader
	var bodyStr string
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
		bodyStr = string(jsonData)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication - prefer token, fallback to basic auth
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	
	req.Header.Set("Content-Type", "application/json")

	log.Printf("AWX API Request: %s %s", method, url)
	if bodyStr != "" {
		log.Printf("Request Body: %s", bodyStr)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("AWX API Response: %d - %s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Detail != "" {
			return fmt.Errorf("AWX API error: %s", errResp.Detail)
		}
		return fmt.Errorf("AWX API error: status %d - %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) GetJobTemplates(ctx context.Context) ([]JobTemplate, error) {
	var response JobTemplateList
	err := c.makeRequest(ctx, "GET", "/api/v2/job_templates/", nil, &response)
	if err != nil {
		return nil, err
	}
	return response.Results, nil
}

func (c *Client) GetJobTemplateByName(ctx context.Context, nameOrID string) (*JobTemplate, error) {
	templates, err := c.GetJobTemplates(ctx)
	if err != nil {
		return nil, err
	}

	// First try to find by exact name match
	for _, template := range templates {
		if template.Name == nameOrID {
			return &template, nil
		}
	}

	// If not found by name, try to parse as ID and find by ID
	if id, err := strconv.Atoi(nameOrID); err == nil {
		for _, template := range templates {
			if template.ID == id {
				return &template, nil
			}
		}
	}

	// If still not found, provide helpful error with available templates
	var availableTemplates []string
	for _, template := range templates {
		availableTemplates = append(availableTemplates, fmt.Sprintf("%s (ID: %d)", template.Name, template.ID))
	}

	return nil, fmt.Errorf("job template '%s' not found. Available templates: %s", nameOrID, strings.Join(availableTemplates, ", "))
}

func (c *Client) LaunchJob(ctx context.Context, templateID int, request LaunchJobRequest) (*JobLaunchResponse, error) {
	var response JobLaunchResponse
	endpoint := fmt.Sprintf("/api/v2/job_templates/%d/launch/", templateID)
	
	// Create a new context without timeout for this specific call
	ctxNoTimeout := context.Background()
	
	// Send minimal request - AWX works fine with empty JSON
	minimalRequest := map[string]interface{}{}
	
	// Only add fields if they have values
	if len(request.ExtraVars) > 0 {
		minimalRequest["extra_vars"] = request.ExtraVars
	}
	if request.Limit != "" {
		minimalRequest["limit"] = request.Limit
	}
	if request.Tags != "" {
		minimalRequest["job_tags"] = request.Tags
	}
	if request.SkipTags != "" {
		minimalRequest["skip_tags"] = request.SkipTags
	}
	
	err := c.makeRequest(ctxNoTimeout, "POST", endpoint, minimalRequest, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) LaunchJobByName(ctx context.Context, templateName string, request LaunchJobRequest) (*JobLaunchResponse, error) {
	template, err := c.GetJobTemplateByName(ctx, templateName)
	if err != nil {
		return nil, err
	}
	return c.LaunchJob(ctx, template.ID, request)
}

func (c *Client) GetJob(ctx context.Context, jobID int) (*Job, error) {
	var job Job
	endpoint := fmt.Sprintf("/api/v2/jobs/%d/", jobID)
	err := c.makeRequest(ctx, "GET", endpoint, nil, &job)
	if err != nil {
		return nil, err
	}
	
	// Set the full URL for the job
	job.URL = c.baseURL + "/#/jobs/playbook/" + strconv.Itoa(jobID)
	
	return &job, nil
}

func (c *Client) TestConnection(ctx context.Context) error {
	log.Printf(" Testing AWX connection to %s", c.baseURL)
	
	// Try to get job templates as a connection test
	_, err := c.GetJobTemplates(ctx)
	if err != nil {
		return fmt.Errorf("AWX connection test failed: %w", err)
	}
	
	log.Printf(" AWX connection test successful")
	return nil
}
