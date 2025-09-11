package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PrometheusClient handles interactions with Prometheus API
type PrometheusClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// PrometheusConfig contains configuration for Prometheus client
type PrometheusConfig struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

// NewPrometheusClient creates a new Prometheus client
func NewPrometheusClient(config PrometheusConfig) *PrometheusClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &PrometheusClient{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		username: config.Username,
		password: config.Password,
	}
}

// QueryResponse represents a Prometheus query response
type QueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string   `json:"resultType"`
		Result     []Result `json:"result"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

// Result represents a single metric result
type Result struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
	Values [][]interface{}   `json:"values,omitempty"`
}

// Query executes a PromQL query
func (c *PrometheusClient) Query(ctx context.Context, query string) (*QueryResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("time", strconv.FormatInt(time.Now().Unix(), 10))

	return c.makeRequest(ctx, "/api/v1/query", params)
}

// QueryRange executes a PromQL range query
func (c *PrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", step.String())

	return c.makeRequest(ctx, "/api/v1/query_range", params)
}

// makeRequest makes a HTTP request to Prometheus API
func (c *PrometheusClient) makeRequest(ctx context.Context, endpoint string, params url.Values) (*QueryResponse, error) {
	reqURL := c.baseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d - %s", resp.StatusCode, string(body))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if queryResp.Status != "success" {
		return nil, fmt.Errorf("query failed: %s - %s", queryResp.ErrorType, queryResp.Error)
	}

	return &queryResp, nil
}

// GetSystemMetrics retrieves common system metrics
func (c *PrometheusClient) GetSystemMetrics(ctx context.Context) (map[string]float64, error) {
	metrics := make(map[string]float64)

	// CPU usage
	cpuQuery := `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	cpuResp, err := c.Query(ctx, cpuQuery)
	if err == nil && len(cpuResp.Data.Result) > 0 {
		if val, err := strconv.ParseFloat(cpuResp.Data.Result[0].Value[1].(string), 64); err == nil {
			metrics["cpu_usage_percent"] = val
		}
	}

	// Memory usage
	memQuery := `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
	memResp, err := c.Query(ctx, memQuery)
	if err == nil && len(memResp.Data.Result) > 0 {
		if val, err := strconv.ParseFloat(memResp.Data.Result[0].Value[1].(string), 64); err == nil {
			metrics["memory_usage_percent"] = val
		}
	}

	// Disk usage
	diskQuery := `100 - ((node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) * 100)`
	diskResp, err := c.Query(ctx, diskQuery)
	if err == nil && len(diskResp.Data.Result) > 0 {
		if val, err := strconv.ParseFloat(diskResp.Data.Result[0].Value[1].(string), 64); err == nil {
			metrics["disk_usage_percent"] = val
		}
	}

	return metrics, nil
}
