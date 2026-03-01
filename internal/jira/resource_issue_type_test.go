package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atlassian/terraform-provider-atlassian/internal/atlassian"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNameMaxLengthValidator_Valid(t *testing.T) {
	v := nameMaxLengthValidator{maxLen: 60}
	req := validator.StringRequest{
		ConfigValue: types.StringValue("Short Name"),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for valid name, got: %v", resp.Diagnostics.Errors())
	}
}

func TestNameMaxLengthValidator_TooLong(t *testing.T) {
	v := nameMaxLengthValidator{maxLen: 60}
	longName := ""
	for i := 0; i < 61; i++ {
		longName += "a"
	}
	req := validator.StringRequest{
		ConfigValue: types.StringValue(longName),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for name exceeding 60 characters, got none")
	}
}

func TestNameMaxLengthValidator_ExactMax(t *testing.T) {
	v := nameMaxLengthValidator{maxLen: 60}
	exactName := ""
	for i := 0; i < 60; i++ {
		exactName += "a"
	}
	req := validator.StringRequest{
		ConfigValue: types.StringValue(exactName),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for name of exactly 60 characters, got: %v", resp.Diagnostics.Errors())
	}
}

func TestNameMaxLengthValidator_Null(t *testing.T) {
	v := nameMaxLengthValidator{maxLen: 60}
	req := validator.StringRequest{
		ConfigValue: types.StringNull(),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null value, got: %v", resp.Diagnostics.Errors())
	}
}

func TestHierarchyLevelValidator_Valid0(t *testing.T) {
	v := hierarchyLevelValidator{}
	req := validator.Int64Request{
		ConfigValue: types.Int64Value(0),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for hierarchy_level 0, got: %v", resp.Diagnostics.Errors())
	}
}

func TestHierarchyLevelValidator_ValidNeg1(t *testing.T) {
	v := hierarchyLevelValidator{}
	req := validator.Int64Request{
		ConfigValue: types.Int64Value(-1),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for hierarchy_level -1, got: %v", resp.Diagnostics.Errors())
	}
}

func TestHierarchyLevelValidator_Invalid(t *testing.T) {
	v := hierarchyLevelValidator{}
	for _, val := range []int64{1, -2, 5, 100} {
		req := validator.Int64Request{
			ConfigValue: types.Int64Value(val),
		}
		resp := &validator.Int64Response{}
		v.ValidateInt64(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("expected error for hierarchy_level %d, got none", val)
		}
	}
}

func TestHierarchyLevelValidator_Null(t *testing.T) {
	v := hierarchyLevelValidator{}
	req := validator.Int64Request{
		ConfigValue: types.Int64Null(),
	}
	resp := &validator.Int64Response{}
	v.ValidateInt64(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for null value, got: %v", resp.Diagnostics.Errors())
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	apiResp := &issueTypeAPIResponse{
		ID:             "10001",
		Name:           "Bug",
		Description:    "A bug type",
		HierarchyLevel: 0,
		AvatarID:       10300,
		IconURL:        "https://example.com/icon.png",
		Subtask:        false,
		Self:           "https://example.atlassian.net/rest/api/3/issuetype/10001",
	}

	model := &IssueTypeResourceModel{}
	mapAPIResponseToState(apiResp, model)

	if model.ID.ValueString() != "10001" {
		t.Errorf("ID = %q, want %q", model.ID.ValueString(), "10001")
	}
	if model.Name.ValueString() != "Bug" {
		t.Errorf("Name = %q, want %q", model.Name.ValueString(), "Bug")
	}
	if model.Description.ValueString() != "A bug type" {
		t.Errorf("Description = %q, want %q", model.Description.ValueString(), "A bug type")
	}
	if model.HierarchyLevel.ValueInt64() != 0 {
		t.Errorf("HierarchyLevel = %d, want %d", model.HierarchyLevel.ValueInt64(), 0)
	}
	if model.AvatarID.ValueInt64() != 10300 {
		t.Errorf("AvatarID = %d, want %d", model.AvatarID.ValueInt64(), 10300)
	}
	if model.IconURL.ValueString() != "https://example.com/icon.png" {
		t.Errorf("IconURL = %q, want %q", model.IconURL.ValueString(), "https://example.com/icon.png")
	}
	if model.Subtask.ValueBool() != false {
		t.Errorf("Subtask = %v, want %v", model.Subtask.ValueBool(), false)
	}
	if model.Self.ValueString() != "https://example.atlassian.net/rest/api/3/issuetype/10001" {
		t.Errorf("Self = %q, want %q", model.Self.ValueString(), "https://example.atlassian.net/rest/api/3/issuetype/10001")
	}
}

func TestMapAPIResponseToState_Subtask(t *testing.T) {
	apiResp := &issueTypeAPIResponse{
		ID:             "10002",
		Name:           "Sub-task",
		Description:    "A subtask type",
		HierarchyLevel: -1,
		AvatarID:       10301,
		IconURL:        "https://example.com/subtask.png",
		Subtask:        true,
		Self:           "https://example.atlassian.net/rest/api/3/issuetype/10002",
	}

	model := &IssueTypeResourceModel{}
	mapAPIResponseToState(apiResp, model)

	if model.HierarchyLevel.ValueInt64() != -1 {
		t.Errorf("HierarchyLevel = %d, want %d", model.HierarchyLevel.ValueInt64(), -1)
	}
	if model.Subtask.ValueBool() != true {
		t.Errorf("Subtask = %v, want %v", model.Subtask.ValueBool(), true)
	}
}

func TestIssueTypeCreateRequest_Serialization(t *testing.T) {
	req := issueTypeCreateRequest{
		Name:           "Bug",
		Description:    "A custom bug",
		HierarchyLevel: 0,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if m["name"] != "Bug" {
		t.Errorf("name = %v, want Bug", m["name"])
	}
	if m["description"] != "A custom bug" {
		t.Errorf("description = %v, want A custom bug", m["description"])
	}
	if m["hierarchyLevel"] != float64(0) {
		t.Errorf("hierarchyLevel = %v, want 0", m["hierarchyLevel"])
	}
}

func TestIssueTypeCreateRequest_OmitEmptyDescription(t *testing.T) {
	req := issueTypeCreateRequest{
		Name:           "Task",
		HierarchyLevel: 0,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, ok := m["description"]; ok {
		t.Error("expected description to be omitted when empty")
	}
}

func TestIssueTypeUpdateRequest_Serialization(t *testing.T) {
	avatarID := int64(10300)
	req := issueTypeUpdateRequest{
		Name:        "Updated Bug",
		Description: "Updated description",
		AvatarID:    &avatarID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if m["name"] != "Updated Bug" {
		t.Errorf("name = %v, want Updated Bug", m["name"])
	}
	if m["description"] != "Updated description" {
		t.Errorf("description = %v, want Updated description", m["description"])
	}
	if m["avatarId"] != float64(10300) {
		t.Errorf("avatarId = %v, want 10300", m["avatarId"])
	}
}

func TestIssueTypeUpdateRequest_OmitEmptyFields(t *testing.T) {
	req := issueTypeUpdateRequest{
		Name: "Just Name",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, ok := m["description"]; ok {
		t.Error("expected description to be omitted when empty")
	}
	if _, ok := m["avatarId"]; ok {
		t.Error("expected avatarId to be omitted when nil")
	}
}

func TestIssueTypeAPIResponse_Deserialization(t *testing.T) {
	body := `{
		"id": "10001",
		"name": "Bug",
		"description": "A bug description",
		"hierarchyLevel": 0,
		"avatarId": 10300,
		"iconUrl": "https://example.com/icon.png",
		"subtask": false,
		"self": "https://example.atlassian.net/rest/api/3/issuetype/10001"
	}`

	var resp issueTypeAPIResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if resp.ID != "10001" {
		t.Errorf("ID = %q, want %q", resp.ID, "10001")
	}
	if resp.Name != "Bug" {
		t.Errorf("Name = %q, want %q", resp.Name, "Bug")
	}
	if resp.Description != "A bug description" {
		t.Errorf("Description = %q, want %q", resp.Description, "A bug description")
	}
	if resp.HierarchyLevel != 0 {
		t.Errorf("HierarchyLevel = %d, want %d", resp.HierarchyLevel, 0)
	}
	if resp.AvatarID != 10300 {
		t.Errorf("AvatarID = %d, want %d", resp.AvatarID, 10300)
	}
	if resp.IconURL != "https://example.com/icon.png" {
		t.Errorf("IconURL = %q, want %q", resp.IconURL, "https://example.com/icon.png")
	}
	if resp.Subtask != false {
		t.Errorf("Subtask = %v, want %v", resp.Subtask, false)
	}
	if resp.Self != "https://example.atlassian.net/rest/api/3/issuetype/10001" {
		t.Errorf("Self = %q, want %q", resp.Self, "https://example.atlassian.net/rest/api/3/issuetype/10001")
	}
}

func TestIssueTypeCreate_HTTPRequest(t *testing.T) {
	var receivedMethod string
	var receivedPath string
	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "10001",
			"name": "Bug",
			"description": "A bug type",
			"hierarchyLevel": 0,
			"avatarId": 10300,
			"iconUrl": "https://example.com/icon.png",
			"subtask": false,
			"self": "https://example.atlassian.net/rest/api/3/issuetype/10001"
		}`))
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	createReq := issueTypeCreateRequest{
		Name:           "Bug",
		Description:    "A bug type",
		HierarchyLevel: 0,
	}
	var apiResp issueTypeAPIResponse
	err := client.Post(context.Background(), "/rest/api/3/issuetype", createReq, &apiResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != "POST" {
		t.Errorf("method = %q, want POST", receivedMethod)
	}
	if receivedPath != "/rest/api/3/issuetype" {
		t.Errorf("path = %q, want /rest/api/3/issuetype", receivedPath)
	}
	if receivedBody["name"] != "Bug" {
		t.Errorf("body name = %v, want Bug", receivedBody["name"])
	}
	if receivedBody["hierarchyLevel"] != float64(0) {
		t.Errorf("body hierarchyLevel = %v, want 0", receivedBody["hierarchyLevel"])
	}
	if apiResp.ID != "10001" {
		t.Errorf("response ID = %q, want 10001", apiResp.ID)
	}
	if apiResp.AvatarID != 10300 {
		t.Errorf("response AvatarID = %d, want 10300", apiResp.AvatarID)
	}
}

func TestIssueTypeRead_HTTPRequest(t *testing.T) {
	var receivedPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "10001",
			"name": "Bug",
			"description": "A bug type",
			"hierarchyLevel": 0,
			"avatarId": 10300,
			"iconUrl": "https://example.com/icon.png",
			"subtask": false,
			"self": "https://example.atlassian.net/rest/api/3/issuetype/10001"
		}`))
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	var apiResp issueTypeAPIResponse
	err := client.Get(context.Background(), "/rest/api/3/issuetype/10001", &apiResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPath != "/rest/api/3/issuetype/10001" {
		t.Errorf("path = %q, want /rest/api/3/issuetype/10001", receivedPath)
	}
	if apiResp.Name != "Bug" {
		t.Errorf("response Name = %q, want Bug", apiResp.Name)
	}
}

func TestIssueTypeRead_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue type not found"]}`))
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	var apiResp issueTypeAPIResponse
	err := client.Get(context.Background(), "/rest/api/3/issuetype/99999", &apiResp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*atlassian.APIError)
	if !ok {
		t.Fatalf("expected *atlassian.APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("status = %d, want 404", apiErr.StatusCode)
	}
}

func TestIssueTypeUpdate_HTTPRequest(t *testing.T) {
	var receivedMethod string
	var receivedPath string
	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "10001",
			"name": "Updated Bug",
			"description": "Updated description",
			"hierarchyLevel": 0,
			"avatarId": 10300,
			"iconUrl": "https://example.com/icon.png",
			"subtask": false,
			"self": "https://example.atlassian.net/rest/api/3/issuetype/10001"
		}`))
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	updateReq := issueTypeUpdateRequest{
		Name:        "Updated Bug",
		Description: "Updated description",
	}
	var apiResp issueTypeAPIResponse
	err := client.Put(context.Background(), "/rest/api/3/issuetype/10001", updateReq, &apiResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != "PUT" {
		t.Errorf("method = %q, want PUT", receivedMethod)
	}
	if receivedPath != "/rest/api/3/issuetype/10001" {
		t.Errorf("path = %q, want /rest/api/3/issuetype/10001", receivedPath)
	}
	if receivedBody["name"] != "Updated Bug" {
		t.Errorf("body name = %v, want Updated Bug", receivedBody["name"])
	}
	if apiResp.Name != "Updated Bug" {
		t.Errorf("response Name = %q, want Updated Bug", apiResp.Name)
	}
}

func TestIssueTypeDelete_HTTPRequest(t *testing.T) {
	var receivedMethod string
	var receivedPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	err := client.Delete(context.Background(), "/rest/api/3/issuetype/10001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != "DELETE" {
		t.Errorf("method = %q, want DELETE", receivedMethod)
	}
	if receivedPath != "/rest/api/3/issuetype/10001" {
		t.Errorf("path = %q, want /rest/api/3/issuetype/10001", receivedPath)
	}
}

func TestIssueTypeSubtaskCreate_HierarchyLevel(t *testing.T) {
	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "10002",
			"name": "Sub-task",
			"description": "",
			"hierarchyLevel": -1,
			"avatarId": 10301,
			"iconUrl": "https://example.com/subtask.png",
			"subtask": true,
			"self": "https://example.atlassian.net/rest/api/3/issuetype/10002"
		}`))
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	createReq := issueTypeCreateRequest{
		Name:           "Sub-task",
		HierarchyLevel: -1,
	}
	var apiResp issueTypeAPIResponse
	err := client.Post(context.Background(), "/rest/api/3/issuetype", createReq, &apiResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["hierarchyLevel"] != float64(-1) {
		t.Errorf("body hierarchyLevel = %v, want -1", receivedBody["hierarchyLevel"])
	}
	if apiResp.Subtask != true {
		t.Errorf("response Subtask = %v, want true", apiResp.Subtask)
	}
	if apiResp.HierarchyLevel != -1 {
		t.Errorf("response HierarchyLevel = %d, want -1", apiResp.HierarchyLevel)
	}
}

func TestIssueTypesDataSource_ListAll(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issuetype" {
			t.Errorf("path = %q, want /rest/api/3/issuetype", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": "10001",
				"name": "Bug",
				"description": "A bug",
				"hierarchyLevel": 0,
				"avatarId": 10300,
				"iconUrl": "https://example.com/bug.png",
				"subtask": false,
				"self": "https://example.atlassian.net/rest/api/3/issuetype/10001"
			},
			{
				"id": "10002",
				"name": "Sub-task",
				"description": "A subtask",
				"hierarchyLevel": -1,
				"avatarId": 10301,
				"iconUrl": "https://example.com/subtask.png",
				"subtask": true,
				"self": "https://example.atlassian.net/rest/api/3/issuetype/10002"
			}
		]`))
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	var apiResp []issueTypeAPIResponse
	err := client.Get(context.Background(), "/rest/api/3/issuetype", &apiResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(apiResp) != 2 {
		t.Fatalf("expected 2 issue types, got %d", len(apiResp))
	}
	if apiResp[0].Name != "Bug" {
		t.Errorf("first issue type name = %q, want Bug", apiResp[0].Name)
	}
	if apiResp[1].Name != "Sub-task" {
		t.Errorf("second issue type name = %q, want Sub-task", apiResp[1].Name)
	}
}

func TestIssueTypesDataSource_FilterByProject(t *testing.T) {
	var receivedPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": "10001",
				"name": "Bug",
				"description": "A bug",
				"hierarchyLevel": 0,
				"avatarId": 10300,
				"iconUrl": "https://example.com/bug.png",
				"subtask": false,
				"self": "https://example.atlassian.net/rest/api/3/issuetype/10001"
			}
		]`))
	}))
	defer ts.Close()

	client := atlassian.NewClient(ts.URL, "user@example.com", "token")
	var apiResp []issueTypeAPIResponse
	err := client.Get(context.Background(), "/rest/api/3/issuetype/project?projectId=10000", &apiResp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPath != "/rest/api/3/issuetype/project?projectId=10000" {
		t.Errorf("path = %q, want /rest/api/3/issuetype/project?projectId=10000", receivedPath)
	}
	if len(apiResp) != 1 {
		t.Fatalf("expected 1 issue type, got %d", len(apiResp))
	}
}
