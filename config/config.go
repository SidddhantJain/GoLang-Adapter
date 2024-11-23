package config

import "os"

// Configuration for API endpoints and base URL.
var (
	BaseURL       = getEnv("BASE_URL", "https://api.definedge.com")    // Default API base URL
	AuthEndpoint  = getEnv("AUTH_ENDPOINT", "/auth/login")            // Endpoint for authentication
	OrderEndpoint = getEnv("ORDER_ENDPOINT", "/placeorder")           // Endpoint for placing orders
)

// getEnv fetches the value of an environment variable or returns a fallback value if not set.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
