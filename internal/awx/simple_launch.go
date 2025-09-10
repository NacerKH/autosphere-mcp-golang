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
	"time"
)

// SimpleLaunchJob - Direct HTTP call without complex context handling
func (c *Client) SimpleLaunchJob(templateID int, extraVars map[string]interface{}) (*JobLaunchResponse, error) {
	url := fmt.Sprintf("%s/api/v2/job_templates/%d/launch/", c.baseURL, templateID)
	
	// Create minimal request body
	requestBody := map[string]interface{}{}
	if len(extraVars) > 0 {
		requestBody["extra_vars"] = extraVars
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	log.Printf("🌐 Simple AWX Launch: POST %s", url)
	log.Printf("📝 Simple Request Body: %s", string(jsonData))
	
	// Create HTTP request with simple timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set basic auth
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")
	
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("📡 Simple AWX Response: %d - %s", resp.StatusCode, string(respBody))
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("AWX error: status %d - %s", resp.StatusCode, string(respBody))
	}
	
	// Parse response
	var response JobLaunchResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &response, nil
}

// SimpleLaunchJobByName - Launch job by template name using simple HTTP
func (c *Client) SimpleLaunchJobByName(templateName string, extraVars map[string]interface{}) (*JobLaunchResponse, error) {
	// Get templates first
	templates, err := c.GetJobTemplates(context.Background())
	if err != nil {
		return nil, err
	}
	
	// Find template by name or ID
	var templateID int
	if id, err := strconv.Atoi(templateName); err == nil {
		// It's an ID
		for _, template := range templates {
			if template.ID == id {
				templateID = id
				break
			}
		}
	} else {
		// It's a name
		for _, template := range templates {
			if template.Name == templateName {
				templateID = template.ID
				break
			}
		}
	}
	
	if templateID == 0 {
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}
	
	return c.SimpleLaunchJob(templateID, extraVars)
}
