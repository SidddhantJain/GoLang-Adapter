package main

import (
	"adapter-project/routes"
	"log"
	"os"
	"github.com/joho/godotenv"
)

func main() {

  err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiToken := os.Getenv("API_TOKEN")
	apiSecret := os.Getenv("API_SECRET")

	if apiToken == "" || apiSecret == "" {
		log.Fatal("API_TOKEN or API_SECRET not set in environment")
	}


	routes.HandleLogin(apiToken, apiSecret)
}
