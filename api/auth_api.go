package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthAPI struct {
	BaseURL string
	Client  *http.Client
}

func NewAuthAPI(baseURL string) *AuthAPI {
	return &AuthAPI{
		BaseURL: "https://integrate.definedgesecurities.com/dart/v1",
		Client:  &http.Client{},
	}
}


func (a *AuthAPI) Login(apiToken, apiSecret string) (*models.LoginResponse, error) {
    payload := map[string]string{
        "api_token":  apiToken,
        "api_secret": apiSecret,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", a.BaseURL+"/login", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := a.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
    }

    var loginResponse models.LoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
        return nil, err
    }

    return &loginResponse, nil
}
