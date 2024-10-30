package config

import "os"

var (
    BaseURL       = getEnv("BASE_URL", "https://api.definedge.com")
    AuthEndpoint  = getEnv("AUTH_ENDPOINT", "/auth/login")
    OrderEndpoint = getEnv("ORDER_ENDPOINT", "/placeorder")
)

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}
