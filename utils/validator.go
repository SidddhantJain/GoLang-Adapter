package utils

func ValidateCredentials(apiToken, apiSecret string) bool {
    return apiToken != "" && apiSecret != ""
}
