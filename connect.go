package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// ConnectToIntegrate struct to hold HTTP client, URLs, and session key
type ConnectToIntegrate struct {
	client        *http.Client
	loginURL      string
	baseURL       string
	apiSessionKey string
}

// NewConnectToIntegrate initializes the ConnectToIntegrate struct
func NewConnectToIntegrate(loginURL, baseURL string) *ConnectToIntegrate {
	return &ConnectToIntegrate{
		client:   &http.Client{},
		loginURL: "https://signin.definedgesecurities.com/auth/realms/debroking/dsbpkc/login/{{api_token}}",
		baseURL:  "https://integrate.definedgesecurities.com/dart/v1",
	}
}

// Login method for ConnectToIntegrate
func (c *ConnectToIntegrate) Login(apiToken, apiSecret string) error {
	// Create the request payload
	payload := map[string]string{
		"api_token":  apiToken,
		"api_secret": apiSecret,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send POST request
	resp, err := c.client.Post(c.loginURL+"/login", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to login")
	}

	// Parse the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Extract session key from the response (assuming it's returned as "api_session_key")
	var loginResponse map[string]interface{}
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		return err
	}

	if key, ok := loginResponse["api_session_key"].(string); ok {
		c.apiSessionKey = key
	} else {
		return errors.New("missing api_session_key")
	}

	return nil
}
