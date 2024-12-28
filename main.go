package main

import (
	"adapter-project/integrate"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	env := godotenv.Load()
	if env != nil {
		print(env)
	}
	connect := integrate.NewConnectToIntegrate("", "", 10, true, nil)
	cerr := connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	if cerr != nil {
		panic(cerr)
	}
}
