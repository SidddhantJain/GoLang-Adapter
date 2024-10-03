package models

// LoginResponse is the response structure for the login API.
type LoginResponse struct {
    APISessionKey string `json:"api_session_key"`
    UID           string `json:"uid"`
    ActID         string `json:"actid"`
}
