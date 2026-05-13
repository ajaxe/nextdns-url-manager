package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient("my-key", "my-profile")
	if client.apiKey != "my-key" {
		t.Errorf("apiKey = %s, want 'my-key'", client.apiKey)
	}
	if client.profileID != "my-profile" {
		t.Errorf("profileID = %s, want 'my-profile'", client.profileID)
	}
	if client.baseURL != "https://api.nextdns.io" {
		t.Errorf("baseURL = %s, want 'https://api.nextdns.io'", client.baseURL)
	}
}

func TestNewAPIClient_EmptyKey(t *testing.T) {
	client := NewAPIClient("", "")
	if client.apiKey != "" {
		t.Errorf("apiKey = %s, want empty", client.apiKey)
	}
	if client.profileID != "" {
		t.Errorf("profileID = %s, want empty", client.profileID)
	}
	if client.baseURL != "https://api.nextdns.io" {
		t.Errorf("baseURL = %s, want default", client.baseURL)
	}
}

func TestSetProfileID(t *testing.T) {
	client := NewAPIClient("key", "old-profile")
	client.SetProfileID("new-profile")
	if client.profileID != "new-profile" {
		t.Errorf("profileID = %s, want 'new-profile'", client.profileID)
	}
}

func TestSetLogChannel(t *testing.T) {
	client := NewAPIClient("key", "profile")
	ch := make(chan string, 10)
	client.SetLogChannel(ch)
	if client.LogChan != ch {
		t.Error("LogChan was not set")
	}
}

func TestSetLogChannel_NilChannel(t *testing.T) {
	client := NewAPIClient("key", "profile")
	var ch chan string
	client.SetLogChannel(ch)
	if client.LogChan != nil {
		t.Error("expected nil LogChan when set with nil")
	}
}

func TestListProfiles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/profiles" {
			t.Errorf("expected path '/profiles', got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "p1", "name": "Profile One"},
				{"id": "p2", "name": "Profile Two"},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient("key", "pid")
	client.baseURL = server.URL

	profiles, err := client.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0]["id"].(string) != "p1" {
		t.Errorf("first profile id = %s, want 'p1'", profiles[0]["id"])
	}
	if profiles[1]["name"].(string) != "Profile Two" {
		t.Errorf("second profile name = %s, want 'Profile Two'", profiles[1]["name"])
	}
}

func TestListProfiles_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewAPIClient("key", "pid")
	client.baseURL = server.URL

	profiles, err := client.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestListProfiles_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer server.Close()

	client := NewAPIClient("key", "pid")
	client.baseURL = server.URL

	profiles, err := client.ListProfiles()
	if err == nil {
		t.Fatal("expected error from API, got nil")
	}
	if profiles != nil {
		t.Errorf("expected nil profiles on error, got %v", profiles)
	}
}

func TestListProfiles_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Internal Server Error"}`))
	}))
	defer server.Close()

	client := NewAPIClient("key", "pid")
	client.baseURL = server.URL

	profiles, err := client.ListProfiles()
	if err == nil {
		t.Fatal("expected error from API, got nil")
	}
}

func TestGetProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "p1",
				"name": "Test Profile",
				"active": true,
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-profile")
	client.baseURL = server.URL

	profile, err := client.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if profile["id"].(string) != "p1" {
		t.Errorf("profile id = %s, want 'p1'", profile["id"])
	}
	if name, ok := profile["name"].(string); !ok || name != "Test Profile" {
		t.Errorf("profile name = %v, want 'Test Profile'", profile["name"])
	}
}

func TestGetProfile_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	profile, err := client.GetProfile()
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
	if profile != nil {
		t.Errorf("expected nil profile, got %v", profile)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Profile not found"}`))
	}))
	defer server.Close()

	client := NewAPIClient("key", "nonexistent")
	client.baseURL = server.URL

	profile, err := client.GetProfile()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if profile != nil {
		t.Errorf("expected nil profile, got %v", profile)
	}
}

func TestAddToDenylist_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/profiles/test-id/denylist" {
			t.Errorf("expected path '/profiles/test-id/denylist', got %s", r.URL.Path)
		}
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if reqBody["id"] != "example.com" {
			t.Errorf("request id = %v, want 'example.com'", reqBody["id"])
		}
		if reqBody["active"] != true {
			t.Errorf("request active = %v, want true", reqBody["active"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	err := client.AddToDenylist("example.com")
	if err != nil {
		t.Fatalf("AddToDenylist() error = %v", err)
	}
}

func TestAddToDenylist_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	err := client.AddToDenylist("example.com")
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
}

func TestAddToDenylist_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"Invalid domain"}`))
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	err := client.AddToDenylist("invalid domain!")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRemoveFromDenylist_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/profiles/test-id/denylist/example.com" {
			t.Errorf("expected path '/profiles/test-id/denylist/example.com', got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	err := client.RemoveFromDenylist("example.com")
	if err != nil {
		t.Fatalf("RemoveFromDenylist() error = %v", err)
	}
}

func TestRemoveFromDenylist_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	err := client.RemoveFromDenylist("example.com")
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
}

func TestListDenylist_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "evil.com", "active": true},
				{"id": "spam.net", "active": true},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	denylist, err := client.ListDenylist()
	if err != nil {
		t.Fatalf("ListDenylist() error = %v", err)
	}
	if len(denylist) != 2 {
		t.Fatalf("expected 2 denylist entries, got %d", len(denylist))
	}
	if denylist[0]["id"].(string) != "evil.com" {
		t.Errorf("first denylist entry = %s, want 'evil.com'", denylist[0]["id"])
	}
}

func TestListDenylist_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	denylist, err := client.ListDenylist()
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
	if denylist != nil {
		t.Errorf("expected nil denylist, got %v", denylist)
	}
}

func TestAddToAllowlist_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/profiles/test-id/allowlist" {
			t.Errorf("expected path '/profiles/test-id/allowlist', got %s", r.URL.Path)
		}
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if reqBody["id"] != "good.com" {
			t.Errorf("request id = %v, want 'good.com'", reqBody["id"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	err := client.AddToAllowlist("good.com")
	if err != nil {
		t.Fatalf("AddToAllowlist() error = %v", err)
	}
}

func TestAddToAllowlist_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	err := client.AddToAllowlist("good.com")
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
}

func TestRemoveFromAllowlist_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/profiles/test-id/allowlist/good.com" {
			t.Errorf("expected path '/profiles/test-id/allowlist/good.com', got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	err := client.RemoveFromAllowlist("good.com")
	if err != nil {
		t.Fatalf("RemoveFromAllowlist() error = %v", err)
	}
}

func TestRemoveFromAllowlist_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	err := client.RemoveFromAllowlist("good.com")
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
}

func TestListAllowlist_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "good.com", "active": true},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	allowlist, err := client.ListAllowlist()
	if err != nil {
		t.Fatalf("ListAllowlist() error = %v", err)
	}
	if len(allowlist) != 1 {
		t.Fatalf("expected 1 allowlist entry, got %d", len(allowlist))
	}
	if allowlist[0]["id"].(string) != "good.com" {
		t.Errorf("allowlist entry = %s, want 'good.com'", allowlist[0]["id"])
	}
}

func TestListAllowlist_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	allowlist, err := client.ListAllowlist()
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
	if allowlist != nil {
		t.Errorf("expected nil allowlist, got %v", allowlist)
	}
}

func TestGetStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profiles/test-id/analytics/status" {
			t.Errorf("expected path '/profiles/test-id/analytics/status', got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "Blocked Domain", "value": 1500},
				{"name": "Allowed Domain", "value": 300},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient("key", "test-id")
	client.baseURL = server.URL

	statuses, err := client.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 status entries, got %d", len(statuses))
	}
}

func TestGetStatus_EmptyProfileID(t *testing.T) {
	client := NewAPIClient("key", "")
	statuses, err := client.GetStatus()
	if err == nil {
		t.Fatal("expected error for empty profile ID")
	}
	if statuses != nil {
		t.Errorf("expected nil status, got %v", statuses)
	}
}

func TestDoRequest_InvalidMethod(t *testing.T) {
	client := NewAPIClient("key", "id")
	_, err := client.doRequest("", "/profiles", nil)
	if err == nil {
		t.Fatal("expected error for empty method")
	}
}

func TestDoRequest_RequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "my-api-key" {
			t.Errorf("X-Api-Key header = %s, want 'my-api-key'", r.Header.Get("X-Api-Key"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want 'application/json'", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient("my-api-key", "pid")
	client.baseURL = server.URL

	_, err := client.doRequest("GET", "/profiles", nil)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}
}

func TestDoRequest_BodySerialization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body["name"] != "test" {
			t.Errorf("body name = %v, want 'test'", body["name"])
		}
		if body["active"] != true {
			t.Errorf("body active = %v, want true", body["active"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient("key", "pid")
	client.baseURL = server.URL

	body := map[string]interface{}{"name": "test", "active": true}
	_, err := client.doRequest("POST", "/profiles/test/denylist", body)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}
}

func TestDoRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	}))
	defer server.Close()

	client := NewAPIClient("key", "pid")
	client.baseURL = server.URL

	_, err := client.doRequest("GET", "/profiles", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDoRequest_NetworkError(t *testing.T) {
	client := NewAPIClient("key", "pid")
	client.baseURL = "http://localhost:99999" // invalid port

	body, err := client.doRequest("GET", "/profiles", nil)
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if body != nil {
		t.Errorf("expected nil body, got %v", body)
	}
}

func TestSyncDisabledApps_EmptyApps(t *testing.T) {
	client := NewAPIClient("key", "pid")
	cfg := &Config{} // no applications

	SyncDisabledApps(client, cfg) // should not panic
}
