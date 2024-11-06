package main

import (
	"adapter-project/services"
	"fmt"
)

func main() {
	authService := services.NewAuthService("https://api.definedgebroking.com/dart/v1/")  //https://www.definedgesecurities.com/

	err := authService.Login("https://signin.definedgesecurities.com/auth/realms/debroking/dsbpkc/login/{{api_token}}",
		"api_secret")
	if err != nil {
		fmt.Println("Login failed:", err)
	} else {
		fmt.Println("Login successful")
	}
}
