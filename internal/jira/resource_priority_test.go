package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Color validation tests ---

func TestStatusColorValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"valid 6-digit hex", "#FF0000", true},
		{"valid 3-digit hex", "#FFF", true},
		{"valid lowercase hex", "#abc123", true},
		{"valid mixed case", "#AbCdEf", true},
		{"valid 3-digit lowercase", "#abc", true},
		{"invalid no hash", "FF0000", false},
		{"invalid 4-digit hex", "#FFFF", false},
		{"invalid 5-digit hex", "#FFFFF", false},
		{"invalid 7-digit hex", "#FFFFFFF", false},
		{"invalid empty", "", false},
		{"invalid just hash", "#", false},
		{"invalid non-hex chars", "#GGGGGG", false},
		{"invalid spaces", "# FF0000", false},
		{"invalid 2-digit hex", "#FF", false},
	}

	v := statusColorValidator{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tt.value),
				Path:        path.Root("status_color"),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			v.ValidateString(ctx, req, resp)

			if tt.isValid && resp.Diagnostics.HasError() {
				t.Errorf("expected color %q to be valid, but got errors: %v", tt.value, resp.Diagnostics.Errors())
			}
			if !tt.isValid && !resp.Diagnostics.HasError() {
				t.Errorf("expected color %q to be invalid, but no errors were reported", tt.value)
			}
		})
	}
}

func TestStatusColorValidator_NullAndUnknown(t *testing.T) {
	v := statusColorValidator{}
	ctx := context.Background()

	// Null value should not produce errors
	req := validator.StringRequest{
		ConfigValue: types.StringNull(),
		Path:        path.Root("status_color"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected null value to pass validation, but got errors")
	}

	// Unknown value should not produce errors
	req = validator.StringRequest{
		ConfigValue: types.StringUnknown(),
		Path:        path.Root("status_color"),
	}
	resp = &validator.StringResponse{}
	v.ValidateString(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected unknown value to pass validation, but got errors")
	}
}

// --- extractTaskPath tests ---

func TestExtractTaskPath(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "full URL",
			url:      "https://site.atlassian.net/rest/api/3/task/10000",
			expected: "/rest/api/3/task/10000",
			wantErr:  false,
		},
		{
			name:     "already a path",
			url:      "/rest/api/3/task/10000",
			expected: "/rest/api/3/task/10000",
			wantErr:  false,
		},
		{
			name:     "URL with port",
			url:      "http://localhost:8080/rest/api/3/task/99999",
			expected: "/rest/api/3/task/99999",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractTaskPath(tt.url)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// --- Async delete with 303 redirect and task polling ---

func TestPriorityResource_AsyncDelete(t *testing.T) {
	taskPollCount := 0

	mux := http.NewServeMux()

	// We need the server URL to construct the Location header, so we use a
	// variable that gets set after server creation.
	var serverURL string

	// DELETE /rest/api/3/priority/10100 -> 303 with Location header
	mux.HandleFunc("/rest/api/3/priority/10100", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Return 303 See Other with task location (full URL like real Jira)
		w.Header().Set("Location", serverURL+"/rest/api/3/task/50000")
		w.WriteHeader(http.StatusSeeOther)
	})

	// GET /rest/api/3/task/50000 -> first call RUNNING, second call COMPLETE
	mux.HandleFunc("/rest/api/3/task/50000", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		taskPollCount++
		var task atlassian.TaskProgress
		if taskPollCount <= 1 {
			task = atlassian.TaskProgress{
				Status:  "RUNNING",
				Message: "Deleting priority...",
			}
		} else {
			task = atlassian.TaskProgress{
				Status:  "COMPLETE",
				Message: "Priority deleted",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	serverURL = server.URL

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	// Simulate the async delete flow
	location, err := client.DeleteWithRedirect(context.Background(), "/rest/api/3/priority/10100")
	if err != nil {
		t.Fatalf("unexpected error from DeleteWithRedirect: %v", err)
	}

	if location == "" {
		t.Fatal("expected non-empty location from 303 redirect")
	}

	taskPath, err := extractTaskPath(location)
	if err != nil {
		t.Fatalf("unexpected error extracting task path: %v", err)
	}

	if taskPath != "/rest/api/3/task/50000" {
		t.Errorf("expected task path /rest/api/3/task/50000, got %s", taskPath)
	}

	err = client.WaitForTask(context.Background(), taskPath, 0)
	if err != nil {
		t.Fatalf("unexpected error waiting for task: %v", err)
	}

	if taskPollCount < 2 {
		t.Errorf("expected at least 2 task polls, got %d", taskPollCount)
	}
}

// --- CRUD tests with httptest ---

func TestPriorityResource_Create(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/priority", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req priorityAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Name != "Critical" {
			t.Errorf("expected name 'Critical', got %s", req.Name)
		}
		if req.StatusColor != "#FF0000" {
			t.Errorf("expected statusColor '#FF0000', got %s", req.StatusColor)
		}
		if req.Description != "Needs immediate attention" {
			t.Errorf("expected description 'Needs immediate attention', got %s", req.Description)
		}

		resp := priorityCreateResponse{ID: "10100"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/rest/api/3/priority/10100", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := priorityAPIResponse{
			ID:          "10100",
			Name:        "Critical",
			Description: "Needs immediate attention",
			StatusColor: "#FF0000",
			IconURL:     "https://example.atlassian.net/images/icons/priorities/critical.svg",
			IsDefault:   false,
			Self:        "https://example.atlassian.net/rest/api/3/priority/10100",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	// Simulate creating the priority through the client
	createReq := priorityAPIRequest{
		Name:        "Critical",
		Description: "Needs immediate attention",
		StatusColor: "#FF0000",
	}

	var createResp priorityCreateResponse
	err := client.Post(context.Background(), "/rest/api/3/priority", createReq, &createResp)
	if err != nil {
		t.Fatalf("unexpected error creating priority: %v", err)
	}

	if createResp.ID != "10100" {
		t.Errorf("expected ID '10100', got %s", createResp.ID)
	}

	// Read back
	var apiResp priorityAPIResponse
	err = client.Get(context.Background(), fmt.Sprintf("/rest/api/3/priority/%s", createResp.ID), &apiResp)
	if err != nil {
		t.Fatalf("unexpected error reading priority: %v", err)
	}

	if apiResp.ID != "10100" {
		t.Errorf("expected ID '10100', got %s", apiResp.ID)
	}
	if apiResp.Name != "Critical" {
		t.Errorf("expected name 'Critical', got %s", apiResp.Name)
	}
	if apiResp.StatusColor != "#FF0000" {
		t.Errorf("expected statusColor '#FF0000', got %s", apiResp.StatusColor)
	}
	if apiResp.IsDefault {
		t.Error("expected isDefault to be false")
	}
}

func TestPriorityResource_Read_NotFound(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/priority/99999", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errorMessages":["Priority not found."]}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	var apiResp priorityAPIResponse
	err := client.Get(context.Background(), "/rest/api/3/priority/99999", &apiResp)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}

	apiErr, ok := err.(*atlassian.APIError)
	if !ok {
		t.Fatalf("expected *atlassian.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestPriorityResource_Update(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/priority/10100", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			var req priorityAPIRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if req.Name != "Updated Critical" {
				t.Errorf("expected name 'Updated Critical', got %s", req.Name)
			}
			if req.StatusColor != "#CC0000" {
				t.Errorf("expected statusColor '#CC0000', got %s", req.StatusColor)
			}

			w.WriteHeader(http.StatusNoContent)

		case http.MethodGet:
			resp := priorityAPIResponse{
				ID:          "10100",
				Name:        "Updated Critical",
				Description: "Updated description",
				StatusColor: "#CC0000",
				IconURL:     "https://example.atlassian.net/images/icons/priorities/critical.svg",
				IsDefault:   false,
				Self:        "https://example.atlassian.net/rest/api/3/priority/10100",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	updateReq := priorityAPIRequest{
		Name:        "Updated Critical",
		Description: "Updated description",
		StatusColor: "#CC0000",
	}

	err := client.Put(context.Background(), "/rest/api/3/priority/10100", updateReq, nil)
	if err != nil {
		t.Fatalf("unexpected error updating priority: %v", err)
	}

	// Read back
	var apiResp priorityAPIResponse
	err = client.Get(context.Background(), "/rest/api/3/priority/10100", &apiResp)
	if err != nil {
		t.Fatalf("unexpected error reading priority: %v", err)
	}

	if apiResp.Name != "Updated Critical" {
		t.Errorf("expected name 'Updated Critical', got %s", apiResp.Name)
	}
	if apiResp.StatusColor != "#CC0000" {
		t.Errorf("expected statusColor '#CC0000', got %s", apiResp.StatusColor)
	}
}

func TestPriorityResource_Delete_TaskFailed(t *testing.T) {
	mux := http.NewServeMux()
	var serverURL string

	mux.HandleFunc("/rest/api/3/priority/10100", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Location", serverURL+"/rest/api/3/task/50001")
		w.WriteHeader(http.StatusSeeOther)
	})

	mux.HandleFunc("/rest/api/3/task/50001", func(w http.ResponseWriter, _ *http.Request) {
		task := atlassian.TaskProgress{
			Status:  "FAILED",
			Message: "Cannot delete the only priority",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	serverURL = server.URL

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	location, err := client.DeleteWithRedirect(context.Background(), "/rest/api/3/priority/10100")
	if err != nil {
		t.Fatalf("unexpected error from DeleteWithRedirect: %v", err)
	}

	taskPath, err := extractTaskPath(location)
	if err != nil {
		t.Fatalf("unexpected error extracting task path: %v", err)
	}

	err = client.WaitForTask(context.Background(), taskPath, 0)
	if err == nil {
		t.Fatal("expected error for failed task, got nil")
	}
}
