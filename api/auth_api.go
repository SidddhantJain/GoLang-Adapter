package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuthAPI handles communication with the authentication API.
type AuthAPI struct {
	BaseURL string       // API base URL
	Client  *http.Client // HTTP client for requests
}

// NewAuthAPI creates a new instance of AuthAPI with the specified base URL.
func NewAuthAPI(baseURL string) *AuthAPI {
	return &AuthAPI{
		BaseURL: "https://integrate.definedgesecurities.com/dart/v1",
		Client:  &http.Client{},
	}
}

// Login authenticates the user with the API using the provided token and secret.
func (a *AuthAPI) Login(apiToken, apiSecret string) (*models.LoginResponse, error) {
	// Create payload with credentials
	payload := map[string]string{
		"api_token":  apiToken,
		"api_secret": apiSecret,
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare login payload: %w", err)
	}

	// Build the HTTP request
	req, err := http.NewRequest("POST", a.BaseURL+"/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request and handle the response
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Decode the response into a LoginResponse struct
	var loginResponse models.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}

	return &loginResponse, nil
}
