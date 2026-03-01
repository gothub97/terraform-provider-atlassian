package jira

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
)

func TestStatusResource_CreateGlobal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/statuses" && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			var req statusCreateRequest
			_ = json.Unmarshal(body, &req)

			if req.Scope.Type != "GLOBAL" {
				t.Errorf("expected scope type GLOBAL, got %s", req.Scope.Type)
			}
			if req.Scope.Project != nil {
				t.Error("expected no project in scope for GLOBAL")
			}
			if len(req.Statuses) != 1 {
				t.Errorf("expected 1 status, got %d", len(req.Statuses))
			}
			if req.Statuses[0].Name != "In Review" {
				t.Errorf("expected name 'In Review', got '%s'", req.Statuses[0].Name)
			}
			if req.Statuses[0].StatusCategory != "IN_PROGRESS" {
				t.Errorf("expected category IN_PROGRESS, got %s", req.Statuses[0].StatusCategory)
			}

			w.Header().Set("Content-Type", "application/json")
			resp := []statusAPIResponse{{
				ID: "10200", Name: "In Review", StatusCategory: "IN_PROGRESS",
				Scope: statusScopeResp{Type: "GLOBAL"},
			}}
			data, _ := json.Marshal(resp)
			_, _ = w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "test@example.com", "token")
	createReq := statusCreateRequest{
		Scope:    statusScope{Type: "GLOBAL"},
		Statuses: []statusCreateEntry{{Name: "In Review", StatusCategory: "IN_PROGRESS"}},
	}
	var created []statusAPIResponse
	err := client.Post(context.Background(), "/rest/api/3/statuses", createReq, &created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 1 || created[0].ID != "10200" {
		t.Errorf("unexpected response: %+v", created)
	}
}

func TestStatusResource_CreateProjectScope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/statuses" && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			var req statusCreateRequest
			_ = json.Unmarshal(body, &req)

			if req.Scope.Type != "PROJECT" {
				t.Errorf("expected scope type PROJECT, got %s", req.Scope.Type)
			}
			if req.Scope.Project == nil || req.Scope.Project.ID != "10001" {
				t.Error("expected project ID 10001 in scope")
			}

			w.Header().Set("Content-Type", "application/json")
			resp := []statusAPIResponse{{
				ID: "10300", Name: "Project Review", StatusCategory: "IN_PROGRESS",
				Scope: statusScopeResp{Type: "PROJECT", Project: &scopeProjectResp{ID: "10001"}},
			}}
			data, _ := json.Marshal(resp)
			_, _ = w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "test@example.com", "token")
	createReq := statusCreateRequest{
		Scope:    statusScope{Type: "PROJECT", Project: &scopeProject{ID: "10001"}},
		Statuses: []statusCreateEntry{{Name: "Project Review", StatusCategory: "IN_PROGRESS"}},
	}
	var created []statusAPIResponse
	err := client.Post(context.Background(), "/rest/api/3/statuses", createReq, &created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 1 || created[0].Scope.Project.ID != "10001" {
		t.Errorf("unexpected response: %+v", created)
	}
}

func TestStatusResource_Validation(t *testing.T) {
	if isValidStatusCategory("INVALID") {
		t.Error("expected INVALID to fail status category validation")
	}
	if !isValidStatusCategory("TODO") {
		t.Error("expected TODO to pass")
	}
	if !isValidStatusCategory("IN_PROGRESS") {
		t.Error("expected IN_PROGRESS to pass")
	}
	if !isValidStatusCategory("DONE") {
		t.Error("expected DONE to pass")
	}
	if isValidScopeType("INVALID") {
		t.Error("expected INVALID to fail scope type validation")
	}
	if !isValidScopeType("GLOBAL") {
		t.Error("expected GLOBAL to pass")
	}
	if !isValidScopeType("PROJECT") {
		t.Error("expected PROJECT to pass")
	}
}

func TestStatusResource_RequestResponseMarshaling(t *testing.T) {
	t.Run("create request marshaling", func(t *testing.T) {
		req := statusCreateRequest{
			Scope:    statusScope{Type: "PROJECT", Project: &scopeProject{ID: "10001"}},
			Statuses: []statusCreateEntry{{Name: "Test", StatusCategory: "TODO", Description: "A test"}},
		}
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var parsed map[string]any
		_ = json.Unmarshal(data, &parsed)
		scope := parsed["scope"].(map[string]any)
		if scope["type"] != "PROJECT" {
			t.Errorf("expected scope type PROJECT, got %v", scope["type"])
		}
		project := scope["project"].(map[string]any)
		if project["id"] != "10001" {
			t.Errorf("expected project id 10001, got %v", project["id"])
		}
	})

	t.Run("global scope omits project", func(t *testing.T) {
		req := statusCreateRequest{
			Scope:    statusScope{Type: "GLOBAL"},
			Statuses: []statusCreateEntry{{Name: "Global", StatusCategory: "DONE"}},
		}
		data, _ := json.Marshal(req)
		var parsed map[string]any
		_ = json.Unmarshal(data, &parsed)
		scope := parsed["scope"].(map[string]any)
		if _, exists := scope["project"]; exists {
			t.Error("GLOBAL scope should not contain project key")
		}
	})

	t.Run("update request marshaling", func(t *testing.T) {
		req := statusUpdateRequest{
			Statuses: []statusUpdateEntry{{ID: "10200", Name: "Updated", StatusCategory: "IN_PROGRESS"}},
		}
		data, _ := json.Marshal(req)
		var parsed map[string]any
		_ = json.Unmarshal(data, &parsed)
		statuses := parsed["statuses"].([]any)
		if len(statuses) != 1 {
			t.Fatalf("expected 1 status, got %d", len(statuses))
		}
		s := statuses[0].(map[string]any)
		if s["id"] != "10200" {
			t.Errorf("expected id 10200, got %v", s["id"])
		}
	})

	t.Run("search response unmarshaling", func(t *testing.T) {
		jsonData := `{"values":[{"id":"10200","name":"In Review","statusCategory":"IN_PROGRESS","scope":{"type":"PROJECT","project":{"id":"10001"}}}]}`
		var resp statusSearchResponse
		if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if len(resp.Values) != 1 || resp.Values[0].ID != "10200" {
			t.Errorf("unexpected response: %+v", resp)
		}
		if resp.Values[0].Scope.Project == nil || resp.Values[0].Scope.Project.ID != "10001" {
			t.Error("expected scope project id 10001")
		}
	})
}
