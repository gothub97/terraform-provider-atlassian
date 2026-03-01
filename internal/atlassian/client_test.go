package atlassian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient_AuthHeader(t *testing.T) {
	var receivedAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token123")
	err := client.Get(context.Background(), "/test", &map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base64 of "user@example.com:token123" is "dXNlckBleGFtcGxlLmNvbTp0b2tlbjEyMw=="
	expected := "Basic dXNlckBleGFtcGxlLmNvbTp0b2tlbjEyMw=="
	if receivedAuth != expected {
		t.Errorf("auth header = %q, want %q", receivedAuth, expected)
	}
}

func TestClient_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"test","value":42}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test" || result.Value != 42 {
		t.Errorf("result = %+v, want {test, 42}", result)
	}
}

func TestClient_Post(t *testing.T) {
	var receivedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type = %s, want application/json", ct)
		}
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"123"}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	body := map[string]string{"key": "TEST"}
	var result struct {
		ID string `json:"id"`
	}
	err := client.Post(context.Background(), "/test", body, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "123" {
		t.Errorf("result.ID = %q, want %q", result.ID, "123")
	}
}

func TestClient_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Not found"]}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	err := client.Get(context.Background(), "/missing", &map[string]any{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("status = %d, want 404", apiErr.StatusCode)
	}
}

func TestClient_Retry429(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := attempts.Add(1)
		if count <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	var result map[string]any
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts.Load() != 3 {
		t.Errorf("attempts = %d, want 3", attempts.Load())
	}
}

func TestClient_Retry5xx(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count := attempts.Add(1)
		if count <= 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	var result map[string]any
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts.Load() != 2 {
		t.Errorf("attempts = %d, want 2", attempts.Load())
	}
}

func TestPaginate(t *testing.T) {
	type Item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		startAt := r.URL.Query().Get("startAt")
		w.Header().Set("Content-Type", "application/json")

		switch startAt {
		case "0", "":
			_, _ = w.Write([]byte(`{
				"startAt": 0, "maxResults": 2, "total": 5, "isLast": false,
				"values": [{"id":"1","name":"a"},{"id":"2","name":"b"}]
			}`))
		case "2":
			_, _ = w.Write([]byte(`{
				"startAt": 2, "maxResults": 2, "total": 5, "isLast": false,
				"values": [{"id":"3","name":"c"},{"id":"4","name":"d"}]
			}`))
		case "4":
			_, _ = w.Write([]byte(`{
				"startAt": 4, "maxResults": 2, "total": 5, "isLast": true,
				"values": [{"id":"5","name":"e"}]
			}`))
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	items, err := Paginate[Item](context.Background(), client, "/items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 5 {
		t.Fatalf("items count = %d, want 5", len(items))
	}
	if items[0].Name != "a" || items[4].Name != "e" {
		t.Errorf("unexpected items: first=%q, last=%q", items[0].Name, items[4].Name)
	}
	if callCount != 3 {
		t.Errorf("page fetches = %d, want 3", callCount)
	}
}

func TestWaitForTask_Complete(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount < 3 {
			_, _ = w.Write([]byte(`{"status":"RUNNING","message":"in progress"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"COMPLETE","message":"done"}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	err := client.WaitForTask(context.Background(), "/task/123", 10*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForTask_Failed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"FAILED","message":"something went wrong"}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	err := client.WaitForTask(context.Background(), "/task/123", 10*time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "task failed: something went wrong" {
		t.Errorf("error = %q, want task failed message", err.Error())
	}
}

func TestDeleteWithRedirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/rest/api/3/task/abc-123")
		w.WriteHeader(http.StatusSeeOther)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user@example.com", "token")
	location, err := client.DeleteWithRedirect(context.Background(), "/priority/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if location != "/rest/api/3/task/abc-123" {
		t.Errorf("location = %q, want /rest/api/3/task/abc-123", location)
	}
}
