package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// APIClient represents a client for interacting with the NextDNS API
type APIClient struct {
	apiKey    string
	profileID string
	baseURL   string
	LogChan   chan string
}

// NewAPIClient creates a new API client with the provided API key
func NewAPIClient(apiKey string, profileID string) *APIClient {
	return &APIClient{
		apiKey:    apiKey,
		profileID: profileID,
		baseURL:   "https://api.nextdns.io",
	}
}

// SetProfileID sets the profile ID for subsequent requests
func (c *APIClient) SetProfileID(profileID string) {
	c.profileID = profileID
}

// SetLogChannel sets the channel for API logging
func (c *APIClient) SetLogChannel(ch chan string) {
	c.LogChan = ch
}

func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonData)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	logMsg := fmt.Sprintf("[DEBUG] %s %s - %s", method, path, resp.Status)
	if c.LogChan != nil {
		select {
		case c.LogChan <- logMsg:
		default:
			// Non-blocking: skip if channel is full
		}
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed: %s - %s - %s", resp.Status, url, string(respBody))
	}

	slog.Debug("API request successful", "method", method, "url", url, "status", resp.Status)

	return respBody, nil
}

// ListProfiles retrieves all profiles associated with the account
func (c *APIClient) ListProfiles() ([]map[string]interface{}, error) {
	data, err := c.doRequest("GET", "/profiles", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetProfile retrieves details for the current profile
func (c *APIClient) GetProfile() (map[string]interface{}, error) {
	if c.profileID == "" {
		return nil, fmt.Errorf("profile ID is required")
	}

	path := fmt.Sprintf("/profiles/%s", c.profileID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// AddToDenylist adds a domain to the denylist
func (c *APIClient) AddToDenylist(domain string) error {
	if c.profileID == "" {
		return fmt.Errorf("profile ID is required")
	}

	body := map[string]interface{}{
		"id":     domain,
		"active": true,
	}

	path := fmt.Sprintf("/profiles/%s/denylist", c.profileID)
	_, err := c.doRequest("POST", path, body)
	return err
}

// RemoveFromDenylist removes a domain from the denylist
func (c *APIClient) RemoveFromDenylist(domain string) error {
	if c.profileID == "" {
		return fmt.Errorf("profile ID is required")
	}

	path := fmt.Sprintf("/profiles/%s/denylist/%s", c.profileID, domain)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// ListDenylist retrieves the denylist for the current profile
func (c *APIClient) ListDenylist() ([]map[string]interface{}, error) {
	if c.profileID == "" {
		return nil, fmt.Errorf("profile ID is required")
	}

	path := fmt.Sprintf("/profiles/%s/denylist", c.profileID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// AddToAllowlist adds a domain to the allowlist
func (c *APIClient) AddToAllowlist(domain string) error {
	if c.profileID == "" {
		return fmt.Errorf("profile ID is required")
	}

	body := map[string]interface{}{
		"id":     domain,
		"active": true,
	}

	path := fmt.Sprintf("/profiles/%s/allowlist", c.profileID)
	_, err := c.doRequest("POST", path, body)
	return err
}

// RemoveFromAllowlist removes a domain from the allowlist
func (c *APIClient) RemoveFromAllowlist(domain string) error {
	if c.profileID == "" {
		return fmt.Errorf("profile ID is required")
	}

	path := fmt.Sprintf("/profiles/%s/allowlist/%s", c.profileID, domain)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// ListAllowlist retrieves the allowlist for the current profile
func (c *APIClient) ListAllowlist() ([]map[string]interface{}, error) {
	if c.profileID == "" {
		return nil, fmt.Errorf("profile ID is required")
	}

	path := fmt.Sprintf("/profiles/%s/allowlist", c.profileID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetStatus retrieves the query status analytics for the current profile
func (c *APIClient) GetStatus() ([]map[string]interface{}, error) {
	if c.profileID == "" {
		return nil, fmt.Errorf("profile ID is required")
	}

	path := fmt.Sprintf("/profiles/%s/analytics/status", c.profileID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}
