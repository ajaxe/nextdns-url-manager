package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock responses
		if r.URL.Path == "/profiles" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "test-id", "name": "Test Profile"},
				},
			})
			return
		}
		if r.Method == "POST" && r.URL.Path == "/profiles/test-id/denylist" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "DELETE" && r.URL.Path == "/profiles/test-id/denylist/example.com" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewAPIClient("fake-key", "test-id")
	client.baseURL = server.URL

	profiles, err := client.ListProfiles()
	if err != nil {
		t.Errorf("ListProfiles failed: %v", err)
	}
	if len(profiles) != 1 || profiles[0]["id"] != "test-id" {
		t.Errorf("Unexpected profiles response")
	}

	err = client.AddToDenylist("example.com")
	if err != nil {
		t.Errorf("AddToDenylist failed: %v", err)
	}

	err = client.RemoveFromDenylist("example.com")
	if err != nil {
		t.Errorf("RemoveFromDenylist failed: %v", err)
	}
}
