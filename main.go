package main

import (
	"adapter-project/integrate"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	env := godotenv.Load()
	if env != nil {
		print(env)
	}
	connect := integrate.NewConnectToIntegrate("", "", 5, true, nil)
	cerr := connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	if cerr != nil {
		panic(cerr)
	}
	orderConnect := integrate.NewIntegrateOrders(connect, true)
	orders, err := orderConnect.Orders()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Orders: %+v\n", orders)

	// gle, err := orderConnect.PlaceOrder("BSE", "BUY", 0, "MARKET", "NORMAL", int(1), "RELIANCE", nil, nil, nil, nil, nil, nil, nil, nil, "DAY")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Order placed: %+v\n", gle)
	// orders, err = orderConnect.Orders()
	// if err != nil {
	// 	panic(err)
	// }
	// // print(connect.APISessionKey)
	// fmt.Printf("Orders: %+v\n", orders)
}
