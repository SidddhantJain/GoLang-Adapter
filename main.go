package api

import (
	"GoLang_adapter/services"
	"fmt"
)

func main() {
	authService := services.NewAuthService("https://www.definedgesecurities.com/")

	err := authService.Login("https://signin.definedgesecurities.com/auth/realms/debroking/dsbpkc/login/{{api_token}}",
		"api_secret")
	if err != nil {
		fmt.Println("Login failed:", err)
	} else {
		fmt.Println("Login successful")
	}
}
