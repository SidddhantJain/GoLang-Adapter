package main

import (
	"adapter-project/integrate"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	env := godotenv.Load()
	if env != nil {
		print(env)
	}
	timed := time.Now()
	connect := integrate.NewConnectToIntegrate("", "", 5, true, nil)
	cerr := connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	if cerr != nil {
		panic(cerr)
	}
	ic := integrate.NewIntegrateDataBeta(connect, true)

	// Example: Get historical data for RELIANCE
	startTime := time.Now().AddDate(0, -1, 0) // 1 month ago
	endTime := time.Now()

	historicalData, err := integrate.HistoricalDataBeta(ic, "NSE", "ACC-EQ", "day", startTime, endTime)
	if err != nil {
		fmt.Println("Error fetching historical data:", err)
	} else {
		for data := range historicalData {
			fmt.Printf("Data: %+v\n", data)
		}
	}
	fmt.Printf("Login successful: %+v\n", time.Since(timed))
	orderConnect := integrate.NewIntegrateOrders(connect, true)
	orders, err := orderConnect.Orders()
	if err != nil {
		panic(err)
	}
	newtc, err := orderConnect.Holdings()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Orders: %+v\n", orders)
	fmt.Printf("New Orders: %+v\n", newtc)

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
