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

// --- Template auto-derivation tests ---

func TestDeriveTemplateKey(t *testing.T) {
	tests := []struct {
		projectTypeKey string
		expected       string
	}{
		{
			projectTypeKey: "software",
			expected:       "com.pyxis.greenhopper.jira:gh-simplified-agility-kanban",
		},
		{
			projectTypeKey: "business",
			expected:       "com.atlassian.jira-core-project-templates:jira-core-simplified-task-tracking",
		},
		{
			projectTypeKey: "service_desk",
			expected:       "com.atlassian.servicedesk:simplified-it-service-management",
		},
		{
			projectTypeKey: "unknown",
			expected:       "",
		},
		{
			projectTypeKey: "",
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.projectTypeKey, func(t *testing.T) {
			result := DeriveTemplateKey(tt.projectTypeKey)
			if result != tt.expected {
				t.Errorf("DeriveTemplateKey(%q) = %q, want %q", tt.projectTypeKey, result, tt.expected)
			}
		})
	}
}

// --- Key validation tests ---

func TestProjectKeyValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"valid short key", "AB", true},
		{"valid long key", "ABCDEFGHIJ", true},
		{"valid with numbers", "ABC123", true},
		{"valid two char", "AZ", true},
		{"valid starts with letter", "Z9", true},
		{"invalid lowercase", "abc", false},
		{"invalid starts with number", "1ABC", false},
		{"invalid single char", "A", false},
		{"invalid too long", "ABCDEFGHIJK", false},
		{"invalid has space", "AB CD", false},
		{"invalid has dash", "AB-CD", false},
		{"invalid has underscore", "AB_CD", false},
		{"invalid empty", "", false},
		{"invalid mixed case", "AbCd", false},
	}

	v := projectKeyValidator{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tt.value),
				Path:        path.Root("key"),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			v.ValidateString(ctx, req, resp)

			if tt.isValid && resp.Diagnostics.HasError() {
				t.Errorf("expected key %q to be valid, but got errors: %v", tt.value, resp.Diagnostics.Errors())
			}
			if !tt.isValid && !resp.Diagnostics.HasError() {
				t.Errorf("expected key %q to be invalid, but no errors were reported", tt.value)
			}
		})
	}
}

func TestProjectKeyValidator_NullAndUnknown(t *testing.T) {
	v := projectKeyValidator{}
	ctx := context.Background()

	// Null value should not produce errors
	req := validator.StringRequest{
		ConfigValue: types.StringNull(),
		Path:        path.Root("key"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected null value to pass validation, but got errors")
	}

	// Unknown value should not produce errors
	req = validator.StringRequest{
		ConfigValue: types.StringUnknown(),
		Path:        path.Root("key"),
	}
	resp = &validator.StringResponse{}
	v.ValidateString(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected unknown value to pass validation, but got errors")
	}
}

// --- Project type key validation tests ---

func TestProjectTypeKeyValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"valid software", "software", true},
		{"valid business", "business", true},
		{"valid service_desk", "service_desk", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown", false},
		{"invalid uppercase", "SOFTWARE", false},
	}

	v := projectTypeKeyValidator{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tt.value),
				Path:        path.Root("project_type_key"),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			v.ValidateString(ctx, req, resp)

			if tt.isValid && resp.Diagnostics.HasError() {
				t.Errorf("expected %q to be valid, but got errors: %v", tt.value, resp.Diagnostics.Errors())
			}
			if !tt.isValid && !resp.Diagnostics.HasError() {
				t.Errorf("expected %q to be invalid, but no errors were reported", tt.value)
			}
		})
	}
}

// --- Assignee type validation tests ---

func TestAssigneeTypeValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"valid PROJECT_LEAD", "PROJECT_LEAD", true},
		{"valid UNASSIGNED", "UNASSIGNED", true},
		{"invalid empty", "", false},
		{"invalid lowercase", "project_lead", false},
		{"invalid unknown", "SOMETHING_ELSE", false},
	}

	v := assigneeTypeValidator{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tt.value),
				Path:        path.Root("assignee_type"),
			}
			resp := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			v.ValidateString(ctx, req, resp)

			if tt.isValid && resp.Diagnostics.HasError() {
				t.Errorf("expected %q to be valid, but got errors: %v", tt.value, resp.Diagnostics.Errors())
			}
			if !tt.isValid && !resp.Diagnostics.HasError() {
				t.Errorf("expected %q to be invalid, but no errors were reported", tt.value)
			}
		})
	}
}

// --- CRUD tests with httptest ---

func TestProjectResource_Create(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/project", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req projectCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Key != "TEST" {
			t.Errorf("expected key TEST, got %s", req.Key)
		}
		if req.Name != "Test Project" {
			t.Errorf("expected name 'Test Project', got %s", req.Name)
		}
		if req.ProjectTypeKey != "software" {
			t.Errorf("expected projectTypeKey 'software', got %s", req.ProjectTypeKey)
		}
		if req.ProjectTemplateKey != "com.pyxis.greenhopper.jira:gh-simplified-agility-kanban" {
			t.Errorf("expected derived template key, got %s", req.ProjectTemplateKey)
		}

		resp := projectCreateResponse{
			ID:   10001,
			Key:  "TEST",
			Self: "https://example.atlassian.net/rest/api/3/project/10001",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/rest/api/3/project/TEST", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := projectAPIResponse{
			ID:             "10001",
			Key:            "TEST",
			Name:           "Test Project",
			ProjectTypeKey: "software",
			Description:    "A test project",
			AssigneeType:   "PROJECT_LEAD",
			Self:           "https://example.atlassian.net/rest/api/3/project/10001",
		}
		resp.Lead.AccountID = "abc123"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	// Simulate creating the project through the client
	createReq := projectCreateRequest{
		Key:                "TEST",
		Name:               "Test Project",
		ProjectTypeKey:     "software",
		ProjectTemplateKey: DeriveTemplateKey("software"),
		LeadAccountID:      "abc123",
		Description:        "A test project",
	}

	var createResp projectCreateResponse
	err := client.Post(context.Background(), "/rest/api/3/project", createReq, &createResp)
	if err != nil {
		t.Fatalf("unexpected error creating project: %v", err)
	}

	if createResp.Key != "TEST" {
		t.Errorf("expected key TEST, got %s", createResp.Key)
	}

	// Read back
	var apiResp projectAPIResponse
	err = client.Get(context.Background(), fmt.Sprintf("/rest/api/3/project/%s", createResp.Key), &apiResp)
	if err != nil {
		t.Fatalf("unexpected error reading project: %v", err)
	}

	if apiResp.ID != "10001" {
		t.Errorf("expected ID 10001, got %s", apiResp.ID)
	}
	if apiResp.Name != "Test Project" {
		t.Errorf("expected name 'Test Project', got %s", apiResp.Name)
	}
	if apiResp.Lead.AccountID != "abc123" {
		t.Errorf("expected lead account ID abc123, got %s", apiResp.Lead.AccountID)
	}
}

func TestProjectResource_Read_NotFound(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/project/99999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errorMessages":["No project could be found with id '99999'."]}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	var apiResp projectAPIResponse
	err := client.Get(context.Background(), "/rest/api/3/project/99999", &apiResp)
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

func TestProjectResource_Update(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/project/10001", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			var req projectUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if req.Name != "Updated Project" {
				t.Errorf("expected name 'Updated Project', got %s", req.Name)
			}

			resp := projectAPIResponse{
				ID:             "10001",
				Key:            "TEST",
				Name:           "Updated Project",
				ProjectTypeKey: "software",
				Description:    "Updated description",
				AssigneeType:   "PROJECT_LEAD",
				Self:           "https://example.atlassian.net/rest/api/3/project/10001",
			}
			resp.Lead.AccountID = "abc123"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case http.MethodGet:
			resp := projectAPIResponse{
				ID:             "10001",
				Key:            "TEST",
				Name:           "Updated Project",
				ProjectTypeKey: "software",
				Description:    "Updated description",
				AssigneeType:   "PROJECT_LEAD",
				Self:           "https://example.atlassian.net/rest/api/3/project/10001",
			}
			resp.Lead.AccountID = "abc123"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	updateReq := projectUpdateRequest{
		Name:          "Updated Project",
		LeadAccountID: "abc123",
		Description:   "Updated description",
	}

	var updateResp projectAPIResponse
	err := client.Put(context.Background(), "/rest/api/3/project/10001", updateReq, &updateResp)
	if err != nil {
		t.Fatalf("unexpected error updating project: %v", err)
	}

	if updateResp.Name != "Updated Project" {
		t.Errorf("expected name 'Updated Project', got %s", updateResp.Name)
	}
}

func TestProjectResource_Delete(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/project/10001", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Verify enableUndo query param
		if r.URL.Query().Get("enableUndo") != "true" {
			t.Error("expected enableUndo=true query parameter")
		}

		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	err := client.Delete(context.Background(), "/rest/api/3/project/10001?enableUndo=true", nil)
	if err != nil {
		t.Fatalf("unexpected error deleting project: %v", err)
	}
}

func TestProjectResource_Delete_NotFound(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/3/project/99999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errorMessages":["No project could be found with id '99999'."]}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := atlassian.NewClient(server.URL, "test@example.com", "test-token")

	err := client.Delete(context.Background(), "/rest/api/3/project/99999?enableUndo=true", nil)
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

func TestMapProjectAPIToState(t *testing.T) {
	apiResp := &projectAPIResponse{
		ID:             "10001",
		Key:            "TEST",
		Name:           "Test Project",
		ProjectTypeKey: "software",
		Description:    "A test project",
		AssigneeType:   "PROJECT_LEAD",
		Self:           "https://example.atlassian.net/rest/api/3/project/10001",
	}
	apiResp.Lead.AccountID = "abc123"

	state := &ProjectResourceModel{}
	mapProjectAPIToState(state, apiResp, "com.pyxis.greenhopper.jira:gh-simplified-agility-kanban")

	if state.ID.ValueString() != "10001" {
		t.Errorf("expected ID '10001', got %q", state.ID.ValueString())
	}
	if state.Key.ValueString() != "TEST" {
		t.Errorf("expected Key 'TEST', got %q", state.Key.ValueString())
	}
	if state.Name.ValueString() != "Test Project" {
		t.Errorf("expected Name 'Test Project', got %q", state.Name.ValueString())
	}
	if state.ProjectTypeKey.ValueString() != "software" {
		t.Errorf("expected ProjectTypeKey 'software', got %q", state.ProjectTypeKey.ValueString())
	}
	if state.ProjectTemplateKey.ValueString() != "com.pyxis.greenhopper.jira:gh-simplified-agility-kanban" {
		t.Errorf("expected ProjectTemplateKey to be set, got %q", state.ProjectTemplateKey.ValueString())
	}
	if state.LeadAccountID.ValueString() != "abc123" {
		t.Errorf("expected LeadAccountID 'abc123', got %q", state.LeadAccountID.ValueString())
	}
	if state.Description.ValueString() != "A test project" {
		t.Errorf("expected Description 'A test project', got %q", state.Description.ValueString())
	}
	if state.AssigneeType.ValueString() != "PROJECT_LEAD" {
		t.Errorf("expected AssigneeType 'PROJECT_LEAD', got %q", state.AssigneeType.ValueString())
	}
	if state.Self.ValueString() != "https://example.atlassian.net/rest/api/3/project/10001" {
		t.Errorf("expected Self URL, got %q", state.Self.ValueString())
	}
}

func TestMapProjectAPIToState_EmptyTemplateKey(t *testing.T) {
	apiResp := &projectAPIResponse{
		ID:             "10001",
		Key:            "TEST",
		Name:           "Test Project",
		ProjectTypeKey: "software",
	}
	apiResp.Lead.AccountID = "abc123"

	state := &ProjectResourceModel{}
	mapProjectAPIToState(state, apiResp, "")

	if !state.ProjectTemplateKey.IsNull() {
		t.Errorf("expected ProjectTemplateKey to be null when templateKey is empty, got %q", state.ProjectTemplateKey.ValueString())
	}
}
