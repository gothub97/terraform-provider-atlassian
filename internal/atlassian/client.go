package atlassian

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxRetries       = 5
	baseRetryDelay   = 1 * time.Second
	maxRetryDelay    = 30 * time.Second
	jitterFraction   = 0.2
	defaultPollDelay = 2 * time.Second
	defaultTimeout   = 2 * time.Minute
)

// Client is the centralized HTTP client for the Atlassian REST API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	authHeader string
}

// NewClient constructs a new Client with Basic Auth credentials.
func NewClient(baseURL, email, apiToken string) *Client {
	cred := base64.StdEncoding.EncodeToString([]byte(email + ":" + apiToken))
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		authHeader: "Basic " + cred,
	}
}

// Get performs a GET request and unmarshals the response into result.
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}


// Post performs a POST request and unmarshals the response into result.
func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request and unmarshals the response into result.
func (c *Client) Put(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request. The result parameter can be nil for 204 responses.
func (c *Client) Delete(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, result)
}

// DeleteWithRedirect performs a DELETE request and returns the redirect Location header.
// Used for async operations like priority delete that return 303.
func (c *Client) DeleteWithRedirect(ctx context.Context, path string) (string, error) {
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return "", err
	}

	// Don't follow redirects automatically
	client := &http.Client{
		Timeout: c.HTTPClient.Timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var resp *http.Response
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err != nil {
			return "", fmt.Errorf("request failed: %w", err)
		}

		if !isRetryable(resp.StatusCode) {
			break
		}
		retryDelay := calculateRetryDelay(resp, attempt)
		_ = resp.Body.Close()

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(retryDelay):
		}

		req, err = c.newRequest(ctx, http.MethodDelete, path, nil)
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("redirect response without Location header")
		}
		// Strip base URL so callers get a relative path (e.g. /rest/api/3/task/123)
		if strings.HasPrefix(location, c.BaseURL) {
			location = strings.TrimPrefix(location, c.BaseURL)
		}
		return location, nil
	}

	if resp.StatusCode == http.StatusNoContent {
		return "", nil
	}

	body, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}

	var resp *http.Response
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = c.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		if !isRetryable(resp.StatusCode) {
			break
		}
		retryDelay := calculateRetryDelay(resp, attempt)
		_ = resp.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryDelay):
		}

		req, err = c.newRequest(ctx, method, path, body)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func isRetryable(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func calculateRetryDelay(resp *http.Response, attempt int) time.Duration {
	if resp.StatusCode == http.StatusTooManyRequests {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	delay := float64(baseRetryDelay) * math.Pow(2, float64(attempt))
	if delay > float64(maxRetryDelay) {
		delay = float64(maxRetryDelay)
	}

	jitter := delay * jitterFraction * (2*rand.Float64() - 1)
	delay += jitter

	return time.Duration(delay)
}

// APIError represents an error response from the Atlassian API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Body)
}

// PaginatedResponse represents the standard Jira pagination envelope.
type PaginatedResponse[T any] struct {
	StartAt    int  `json:"startAt"`
	MaxResults int  `json:"maxResults"`
	Total      int  `json:"total"`
	IsLast     bool `json:"isLast"`
	Values     []T  `json:"values"`
}

// Paginate fetches all pages of a paginated endpoint and returns the accumulated results.
func Paginate[T any](ctx context.Context, c *Client, basePath string) ([]T, error) {
	var allValues []T
	startAt := 0

	for {
		separator := "?"
		if len(basePath) > 0 {
			for _, ch := range basePath {
				if ch == '?' {
					separator = "&"
					break
				}
			}
		}
		path := fmt.Sprintf("%s%sstartAt=%d&maxResults=50", basePath, separator, startAt)

		var page PaginatedResponse[T]
		if err := c.Get(ctx, path, &page); err != nil {
			return nil, fmt.Errorf("fetching page at startAt=%d: %w", startAt, err)
		}

		allValues = append(allValues, page.Values...)

		if page.IsLast || startAt+page.MaxResults >= page.Total {
			break
		}
		startAt += page.MaxResults
	}

	return allValues, nil
}

// TaskProgress represents the response from GET /rest/api/3/task/{taskId}.
type TaskProgress struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// WaitForTask polls a task endpoint until it reaches COMPLETE or FAILED status.
func (c *Client) WaitForTask(ctx context.Context, taskPath string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("task timed out after %v", timeout)
		}

		var task TaskProgress
		if err := c.Get(ctx, taskPath, &task); err != nil {
			return fmt.Errorf("polling task: %w", err)
		}

		switch task.Status {
		case "COMPLETE":
			return nil
		case "FAILED":
			return fmt.Errorf("task failed: %s", task.Message)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultPollDelay):
		}
	}
}
