// Package confluence contains the Terraform resources and data sources for
// managing Atlassian Confluence Cloud spaces and space permissions.
//
// It reuses the shared *atlassian.Client (Basic auth, retry/backoff, APIError)
// from internal/atlassian. Confluence endpoints live under the "/wiki" prefix:
//   - v2 REST API:  /wiki/api/v2/...      (space create/read, role assignments)
//   - v1 REST API:  /wiki/rest/api/...    (space update/delete, granular permissions)
package confluence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
)

const (
	// v2 and v1 base paths for Confluence Cloud.
	apiV2Base = "/wiki/api/v2"
	apiV1Base = "/wiki/rest/api"

	// longTaskPollDelay is how long to wait between polls of a v1 long-running task.
	longTaskPollDelay = 2 * time.Second
	// longTaskTimeout bounds the wait for an async operation (e.g. space delete).
	longTaskTimeout = 2 * time.Minute
)

// configureClient is the shared Configure boilerplate for Confluence resources and
// data sources. It extracts the *atlassian.Client injected by the provider.
//
// Returns (client, ok). When ok is false the caller should return immediately;
// errMsg describes the problem for diagnostics (empty when providerData is nil,
// which is a normal pre-configuration call and not an error).
func configureClient(providerData any) (client *atlassian.Client, errMsg string, ok bool) {
	if providerData == nil {
		return nil, "", false
	}
	c, isClient := providerData.(*atlassian.Client)
	if !isClient {
		return nil, fmt.Sprintf("Expected *atlassian.Client, got: %T. Please report this issue to the provider developers.", providerData), false
	}
	return c, "", true
}

// splitCompositeID splits a composite import ID (e.g. "SPACEKEY/group-id") into
// exactly n parts, erroring if the shape does not match.
func splitCompositeID(id string, n int) ([]string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != n {
		return nil, fmt.Errorf("expected import ID in the form %q, got %q",
			strings.Join(placeholders(n), "/"), id)
	}
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("import ID segment %d is empty in %q", i+1, id)
		}
	}
	return parts, nil
}

func placeholders(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("<part%d>", i+1)
	}
	return out
}

// longTaskResponse is the body returned by a v1 async operation (202 Accepted),
// pointing at the long-running task status endpoint.
type longTaskResponse struct {
	ID    string `json:"id"`
	Links struct {
		Status string `json:"status"`
	} `json:"links"`
}

// longTaskStatus is the v1 GET /wiki/rest/api/longtask/{id} response.
type longTaskStatus struct {
	ID                 string `json:"id"`
	Finished           bool   `json:"finished"`
	Successful         bool   `json:"successful"`
	PercentageComplete int    `json:"percentageComplete"`
	Messages           []struct {
		Translation string `json:"translation"`
	} `json:"messages"`
}

// waitForLongTask polls a Confluence v1 long-running task until it finishes.
// taskID is the long task identifier (as returned in longTaskResponse.ID). It
// returns nil on successful completion, or an error on failure/timeout.
func waitForLongTask(ctx context.Context, client *atlassian.Client, taskID string) error {
	path := fmt.Sprintf("%s/longtask/%s", apiV1Base, taskID)
	deadline := time.Now().Add(longTaskTimeout)

	for {
		var status longTaskStatus
		if err := client.Get(ctx, path, &status); err != nil {
			return fmt.Errorf("polling long task %s: %w", taskID, err)
		}
		if status.Finished {
			if status.Successful {
				return nil
			}
			msg := "long task did not complete successfully"
			if len(status.Messages) > 0 {
				msg = status.Messages[0].Translation
			}
			return fmt.Errorf("long task %s failed: %s", taskID, msg)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("long task %s timed out after %v", taskID, longTaskTimeout)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(longTaskPollDelay):
		}
	}
}
