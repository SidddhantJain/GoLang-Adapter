package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "adapter-project/models"
)

func Login(apiToken, apiSecret string) (*models.LoginResponse, error) {
    payload := map[string]string{
        "api_token":  apiToken,
        "api_secret": apiSecret,
    }
    jsonPayload, _ := json.Marshal(payload)
    resp, err := http.Post("https://api.definedge.com/auth/login", "application/json", bytes.NewBuffer(jsonPayload))
    
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var loginResp models.LoginResponse
    json.NewDecoder(resp.Body).Decode(&loginResp)
    
    return &loginResp, nil
}
