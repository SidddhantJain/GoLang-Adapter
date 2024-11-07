package main

import (
	"adapter-project/integrate"
)

func main() {
	connect := integrate.NewConnectToIntegrate("", "", 10, true, nil)
	cerr := connect.Login()
}
