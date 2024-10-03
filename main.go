package main

import (
    "adapter-project/routes"
    "adapter-project/config"
    "log"
    "os"
)

func main() {
    // Load environment variables
    apiToken := os.Getenv("API_TOKEN")
    apiSecret := os.Getenv("API_SECRET")
    
    if apiToken == "" || apiSecret == "" {
        log.Fatal("API_TOKEN or API_SECRET not set in environment")
    }
    
    // Call route handler to log in
    routes.HandleLogin(apiToken, apiSecret)
}
