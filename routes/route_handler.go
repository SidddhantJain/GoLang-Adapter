package routes

import (
    "adapter-project/services"
    "log"
)

func HandleLogin(apiToken, apiSecret string) {
    loginResponse, err := services.LoginService(apiToken, apiSecret)
    if err != nil {
        log.Fatalf("Login failed: %v", err)
    }
    log.Printf("Login successful, session key: %s", loginResponse.APISessionKey)
}
