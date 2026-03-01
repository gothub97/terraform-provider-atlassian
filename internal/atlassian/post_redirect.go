package atlassian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PostWithRedirect performs a POST request and returns the redirect Location header.
// Used for async operations like workflow scheme draft publish that return 303.
func (c *Client) PostWithRedirect(ctx context.Context, path string, body any) (string, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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

		// Rebuild the request for retry
		if body != nil {
			jsonBody, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				return "", fmt.Errorf("marshaling request body: %w", marshalErr)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
		if err != nil {
			return "", fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", c.authHeader)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("redirect response without Location header")
		}
		return location, nil
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "", nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return "", &APIError{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
	}
}
