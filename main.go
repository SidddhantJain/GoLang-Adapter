package main

import (
	"fmt"
	"os"
	"pyintegrate/integrate"
	"pyintegrate/structs"
	"time"

	"github.com/joho/godotenv"
)

func TestHistoricalDataBeta() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	ic := integrate.NewIntegrateDataBeta(connect, false)
	startTime := time.Now().AddDate(0, -1, 0)
	endTime := time.Now()
	ch, err := integrate.HistoricalDataBeta(ic, "NSE", "ACC-EQ", "day", startTime, endTime)
	if err != nil {
		return fmt.Errorf("HistoricalDataBeta error: %v", err)
	}
	count := 0
	for data := range ch {
		if data["open"] == nil {
			return fmt.Errorf("Missing open field in historical data")
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("No historical data returned")
	}
	return nil
}

func TestPlaceOrder() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.PlaceOrder(
		"NSE", "BUY", 0, "MARKET", "INTRADAY", 1, "ACC-EQ",
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if err != nil {
		return fmt.Errorf("PlaceOrder error: %v", err)
	}
	if resp["order_id"] == nil {
		return fmt.Errorf("No order_id returned in response")
	}
	return nil
}

func TestQuotesBeta() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	ic := integrate.NewIntegrateDataBeta(connect, false)
	resp, err := integrate.QuotesBeta(ic, "NSE", "ACC-EQ")
	if err != nil {
		return fmt.Errorf("QuotesBeta error: %v", err)
	}
	if resp["ltp"] == nil {
		return fmt.Errorf("No ltp returned in quotes response")
	}
	return nil
}

func TestOrdersEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.Orders()
	if err != nil {
		return fmt.Errorf("Orders endpoint error: %v", err)
	}
	orders, ok := resp["orders"].([]interface{})
	if !ok {
		return fmt.Errorf("No orders field in response or wrong type")
	}
	if len(orders) == 0 {
		return fmt.Errorf("No orders returned in response")
	}
	for _, o := range orders {
		order, ok := o.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Order item is not a map")
		}
		if order["order_id"] == nil {
			return fmt.Errorf("Order missing order_id field")
		}
		if order["tradingsymbol"] == nil {
			return fmt.Errorf("Order missing tradingsymbol field")
		}
	}
	return nil
}

func TestTradeBookEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.Trades()
	if err != nil {
		return fmt.Errorf("Trade Book endpoint error: %v", err)
	}
	trades, ok := resp["trades"].([]interface{})
	if !ok {
		return fmt.Errorf("No trades field in response or wrong type")
	}
	for _, tr := range trades {
		trade, ok := tr.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Trade item is not a map")
		}
		if trade["order_id"] == nil {
			return fmt.Errorf("Trade missing order_id field")
		}
		if trade["tradingsymbol"] == nil {
			return fmt.Errorf("Trade missing tradingsymbol field")
		}
	}
	return nil
}

func TestPositionBookEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.Positions()
	if err != nil {
		return fmt.Errorf("Position Book endpoint error: %v", err)
	}
	positions, ok := resp["positions"].([]interface{})
	if !ok {
		return fmt.Errorf("No positions field in response or wrong type")
	}
	for _, pos := range positions {
		position, ok := pos.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Position item is not a map")
		}
		if position["tradingsymbol"] == nil {
			return fmt.Errorf("Position missing tradingsymbol field")
		}
	}
	return nil
}

func TestHoldingsEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.Holdings()
	if err != nil {
		return fmt.Errorf("Holdings endpoint error: %v", err)
	}
	data, ok := resp["data"].([]interface{})
	if !ok {
		return fmt.Errorf("No data field in holdings response or wrong type")
	}
	for _, h := range data {
		holding, ok := h.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Holding item is not a map")
		}
		if holding["tradingsymbol"] == nil {
			return fmt.Errorf("Holding missing tradingsymbol field")
		}
	}
	return nil
}

func TestModifyOrderEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	params := structs.ModifyOrderParams{
		Exchange:      "NFO",
		OrderID:       "1234567890012",
		OrderType:     "BUY",
		Price:         220,
		PriceType:     "LIMIT",
		ProductType:   "INTRADAY",
		Quantity:      100,
		TradingSymbol: "NIFTY23FEB23C17800",
		Validity:      "DAY",
	}
	resp, err := orderConnect.ModifyOrder(params)
	if err != nil {
		return fmt.Errorf("Modify Order endpoint error: %v", err)
	}
	if resp["order_id"] == nil {
		return fmt.Errorf("No order_id returned in modify order response")
	}
	return nil
}

func TestCancelOrderEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.CancelOrder("1234567890012")
	if err != nil {
		return fmt.Errorf("Cancel Order endpoint error: %v", err)
	}
	if resp["order_id"] == nil {
		return fmt.Errorf("No order_id returned in cancel order response")
	}
	return nil
}

func TestSliceOrderEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.SliceOrder(
		"NFO", "BUY", 17600, "LIMIT", "NORMAL", 1000, 2, "NIFTY23FEB23F",
		nil, nil, nil, nil, nil, nil, nil, nil, "DAY",
	)
	if err != nil {
		return fmt.Errorf("Slice Order endpoint error: %v", err)
	}
	orders, ok := resp["orders"].([]interface{})
	if !ok {
		return fmt.Errorf("No orders field in slice order response or wrong type")
	}
	for _, o := range orders {
		order, ok := o.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Slice order item is not a map")
		}
		if order["order_id"] == nil {
			return fmt.Errorf("Slice order missing order_id field")
		}
	}
	return nil
}

func TestProductConversionEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.ConvertPositionProductType(
		"NFO", "SELL", "NORMAL", "INTRADAY", 50, "NIFTY23FEB23F", "DAY",
	)
	if err != nil {
		return fmt.Errorf("Product Conversion endpoint error: %v", err)
	}
	if resp["status"] != "SUCCESS" {
		return fmt.Errorf("Product conversion did not return SUCCESS status")
	}
	return nil
}

func TestGTTOrderBookEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.GTTOrders()
	if err != nil {
		return fmt.Errorf("GTT Order Book endpoint error: %v", err)
	}
	pendingGTTOrderBook, ok := resp["pendingGTTOrderBook"].([]interface{})
	if !ok {
		return fmt.Errorf("No pendingGTTOrderBook field in GTT order book response or wrong type")
	}
	for _, g := range pendingGTTOrderBook {
		gtt, ok := g.(map[string]interface{})
		if !ok {
			return fmt.Errorf("GTT order item is not a map")
		}
		if gtt["alert_id"] == nil {
			return fmt.Errorf("GTT order missing alert_id field")
		}
		if gtt["tradingsymbol"] == nil {
			return fmt.Errorf("GTT order missing tradingsymbol field")
		}
	}
	return nil
}

func TestPlaceGTTOrderEndpoint() error {
	connect := integrate.NewConnectToIntegrate("", "", 5, false, nil)
	_ = connect.Login(os.Getenv("api_token"), os.Getenv("api_secret"), nil)
	orderConnect := integrate.NewIntegrateOrders(connect, false)
	resp, err := orderConnect.PlaceGTTOrder(
		"NSE", "BUY", 3100, 1, "TCS-EQ", 3100, "LTP_BELOW",
	)
	if err != nil {
		return fmt.Errorf("Place GTT Order endpoint error: %v", err)
	}
	if resp["alert_id"] == nil {
		return fmt.Errorf("No alert_id returned in place GTT order response")
	}
	return nil
}

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

	// Run all tests and print results
	tests := []struct {
		name string
		test func() error
	}{
		{"TestHistoricalDataBeta", TestHistoricalDataBeta},
		{"TestPlaceOrder", TestPlaceOrder},
		{"TestQuotesBeta", TestQuotesBeta},
		{"TestOrdersEndpoint", TestOrdersEndpoint},
		{"TestTradeBookEndpoint", TestTradeBookEndpoint},
		{"TestPositionBookEndpoint", TestPositionBookEndpoint},
		{"TestHoldingsEndpoint", TestHoldingsEndpoint},
		{"TestModifyOrderEndpoint", TestModifyOrderEndpoint},
		{"TestCancelOrderEndpoint", TestCancelOrderEndpoint},
		{"TestSliceOrderEndpoint", TestSliceOrderEndpoint},
		{"TestProductConversionEndpoint", TestProductConversionEndpoint},
		{"TestGTTOrderBookEndpoint", TestGTTOrderBookEndpoint},
		{"TestPlaceGTTOrderEndpoint", TestPlaceGTTOrderEndpoint},
	}
	fmt.Println("Running API tests:")
	for _, tc := range tests {
		err := tc.test()
		if err != nil {
			fmt.Printf("%s: FAILED - %v\n", tc.name, err)
		} else {
			fmt.Printf("%s: PASSED\n", tc.name)
		}
	}
}
