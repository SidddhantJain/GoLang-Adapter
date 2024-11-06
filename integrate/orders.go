package integrate

import (
	"adapter-project/structs"
	"errors"
	"fmt"
)

type IntegrateOrders struct {
	c2i     *LocalConnect
	logging bool
}

// NewIntegrateOrders initializes a new instance of IntegrateOrders
func NewIntegrateOrders(connectToIntegrate *structs.ConnectToIntegrate, logging bool) *IntegrateOrders {
	return &IntegrateOrders{
		logging: logging,
		c2i: &LocalConnect{
			ConnectToIntegrate: connectToIntegrate},
	}
}

// PlaceOrder places an order and returns order details.
func (io *IntegrateOrders) PlaceOrder(
	exchange string,
	orderType string,
	price float64,
	priceType string,
	productType string,
	quantity int,
	tradingSymbol string,
	amo *string,
	bookLossPrice *float64,
	bookProfitPrice *float64,
	disclosedQuantity *int,
	marketProtection *float64,
	remarks *string,
	trailingPrice *float64,
	triggerPrice *float64,
	validity string,
) (map[string]interface{}, error) {

	// Validate exchange, order type, price type, and product type
	if !io.isValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !io.isValidOrderType(orderType) {
		return nil, errors.New("invalid order type")
	}
	if !io.isValidPriceType(priceType) {
		return nil, errors.New("invalid price type")
	}
	if !io.isValidProductType(productType) {
		return nil, errors.New("invalid product type")
	}

	// Validate price for market orders
	if priceType == "MARKET" && price != 0 {
		return nil, errors.New("price should be 0 for market order")
	}

	// Validate trigger price for SL-LIMIT orders
	if priceType == "SL-LIMIT" {
		if orderType == "BUY" && triggerPrice != nil && *triggerPrice > price {
			return nil, errors.New("trigger price cannot be greater than price for SL-LIMIT BUY order")
		} else if orderType == "SELL" && triggerPrice != nil && *triggerPrice < price {
			return nil, errors.New("trigger price cannot be lesser than price for SL-LIMIT SELL order")
		}
	}

	// Validate quantity
	if quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}

	// Construct the JSON payload for the request
	jsonParams := map[string]interface{}{
		"exchange":      exchange,
		"order_type":    orderType,
		"price":         price,
		"price_type":    priceType,
		"product_type":  productType,
		"quantity":      quantity,
		"tradingsymbol": tradingSymbol,
		"validity":      validity,
	}

	if amo != nil {
		jsonParams["amo"] = *amo
	}
	if bookLossPrice != nil {
		jsonParams["book_loss_price"] = *bookLossPrice
	}
	if bookProfitPrice != nil {
		jsonParams["book_profit_price"] = *bookProfitPrice
	}
	if disclosedQuantity != nil {
		jsonParams["disclosed_quantity"] = *disclosedQuantity
	}
	if marketProtection != nil {
		jsonParams["market_protection"] = *marketProtection
	}
	if remarks != nil {
		jsonParams["remarks"] = *remarks
	}
	if trailingPrice != nil {
		jsonParams["trailing_price"] = *trailingPrice
	}
	if triggerPrice != nil {
		jsonParams["trigger_price"] = *triggerPrice
	}

	// Send request
	return io.c2i.sendRequest(io.c2i.ConnectToIntegrate.BaseURL, "placeorder", "POST", nil, jsonParams, nil, nil, nil)
}

// Additional helper functions to validate fields

func (io *IntegrateOrders) isValidExchange(exchange string) bool {
	for _, ex := range io.c2i.ConnectToIntegrate.ExchangeTypes {
		if ex == exchange {
			return true
		}
	}
	return false
}

func (io *IntegrateOrders) isValidOrderType(orderType string) bool {
	for _, ot := range io.c2i.ConnectToIntegrate.OrderTypes {
		if ot == orderType {
			return true
		}
	}
	return false
}

func (io *IntegrateOrders) isValidPriceType(priceType string) bool {
	for _, pt := range io.c2i.ConnectToIntegrate.PriceTypes {
		if pt == priceType {
			return true
		}
	}
	return false
}

func (io *IntegrateOrders) isValidProductType(productType string) bool {
	for _, pt := range io.c2i.ConnectToIntegrate.ProductTypes {
		if pt == productType {
			return true
		}
	}
	return false
}

// OrderParams represents the parameters required to modify an order.

// ModifyOrder modifies an open order based on the given parameters.
func (c *IntegrateOrders) ModifyOrder(params structs.ModifyOrderParams) (map[string]interface{}, error) {
	// Check exchange type
	// if !contains(c.c2i.ExchangeTypes, params.Exchange) {
	// 	return nil, errors.New("invalid exchange type")
	// }

	if !c.isValidExchange(params.Exchange) {
		return nil, errors.New("invalid exchange type")
	}

	// Check order type
	// if !contains(c.OrderTypes, params.OrderType) {
	// 	return nil, errors.New("invalid order type")
	// }

	if !c.isValidOrderType(params.OrderType) {
		return nil, errors.New("invalid order type")
	}

	// Check price type
	// if !contains(c.PriceTypes, params.PriceType) {
	// 	return nil, errors.New("invalid price type")
	// }

	if !c.isValidPriceType(params.PriceType) {
		return nil, errors.New("invalid price type")
	}

	// Check product type
	// if !contains(c.ProductTypes, params.ProductType) {
	// 	return nil, errors.New("invalid product type")
	// }

	if !c.isValidProductType(params.ProductType) {
		return nil, errors.New("invalid product type")
	}

	// Validate price for MARKET order type
	if params.PriceType == "MARKET" && params.Price != 0 {
		return nil, errors.New("price should be 0 for market order")
	}

	// Validate SL-LIMIT specific conditions
	if params.PriceType == "SL-LIMIT" {
		if params.OrderType == "BUY" && params.TriggerPrice != nil && *params.TriggerPrice > params.Price {
			return nil, errors.New("trigger price cannot be greater than price for SL-LIMIT BUY order")
		} else if params.OrderType == "SELL" && params.TriggerPrice != nil && *params.TriggerPrice < params.Price {
			return nil, errors.New("trigger price cannot be lesser than price for SL-LIMIT SELL order")
		}
	}

	// Check if quantity is non-zero
	if params.Quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}

	// Prepare JSON parameters for the request
	jsonParams := make(map[string]interface{})
	// Add all fields that are not nil or default values
	addField(jsonParams, "exchange", params.Exchange)
	addField(jsonParams, "order_id", params.OrderID)
	addField(jsonParams, "order_type", params.OrderType)
	addField(jsonParams, "price", params.Price)
	addField(jsonParams, "price_type", params.PriceType)
	addField(jsonParams, "product_type", params.ProductType)
	addField(jsonParams, "quantity", params.Quantity)
	addField(jsonParams, "tradingsymbol", params.TradingSymbol)
	addOptionalField(jsonParams, "amo", params.Amo)
	addOptionalField(jsonParams, "book_loss_price", params.BookLossPrice)
	addOptionalField(jsonParams, "book_profit_price", params.BookProfitPrice)
	addOptionalField(jsonParams, "disclosed_quantity", params.DisclosedQuantity)
	addOptionalField(jsonParams, "market_protection", params.MarketProtection)
	addOptionalField(jsonParams, "remarks", params.Remarks)
	addOptionalField(jsonParams, "trailing_price", params.TrailingPrice)
	addOptionalField(jsonParams, "trigger_price", params.TriggerPrice)
	addField(jsonParams, "validity", params.Validity)

	// Send request
	response, err := c.c2i.sendRequest(c.c2i.BaseURL, "modify", "POST", nil, jsonParams, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to modify order: %w", err)
	}

	return response, nil
}

// Helper function to check if a value exists in a slice.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Helper function to add a non-nil field to the JSON params.
func addField(params map[string]interface{}, key string, value interface{}) {
	params[key] = value
}

// Helper function to add an optional field to the JSON params.
func addOptionalField(params map[string]interface{}, key string, value interface{}) {
	if value != nil {
		params[key] = value
	}
}

// CancelOrder cancels an order based on the order ID.
func (c *IntegrateOrders) CancelOrder(orderID string) (map[string]interface{}, error) {
	if orderID == "" {
		return nil, errors.New("order ID cannot be empty")
	}

	// Prepare the route for cancellation
	route := fmt.Sprintf("cancel/%s", orderID)
	urlParams := map[string]string{"order_id": orderID}
	// Send GET request for cancellation
	response, err := c.c2i.sendRequest(c.c2i.BaseURL, route, "GET", urlParams, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// SliceOrder slices an order into multiple parts and places each as a separate order.
func (c *IntegrateOrders) SliceOrder(
	exchange string,
	orderType string,
	price float64,
	priceType string,
	productType string,
	quantity int,
	slices int,
	tradingsymbol string,
	amo *string,
	bookLossPrice *float64,
	bookProfitPrice *float64,
	disclosedQuantity *int,
	marketProtection *float64,
	remarks *string,
	trailingPrice *float64,
	triggerPrice *float64,
	validity string,
) (map[string]interface{}, error) {

	// Validations
	// if !isValidExchange(exchange) {
	// 	return nil, errors.New("invalid exchange type")
	// }
	// if !isValidOrderType(orderType) {
	// 	return nil, errors.New("invalid order type")
	// }
	// if !isValidPriceType(priceType) {
	// 	return nil, errors.New("invalid price type")
	// }
	// if !isValidProductType(productType) {
	// 	return nil, errors.New("invalid product type")
	// }

	if !c.isValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !c.isValidOrderType(orderType) {
		return nil, errors.New("invalid order type")
	}
	if !c.isValidPriceType(priceType) {
		return nil, errors.New("invalid price type")
	}
	if !c.isValidProductType(productType) {
		return nil, errors.New("invalid product type")
	}

	if priceType == "MARKET" && price != 0 {
		return nil, errors.New("price should be 0 for market order")
	}
	if priceType == "SL-LIMIT" {
		if (orderType == "BUY" && triggerPrice != nil && *triggerPrice > price) ||
			(orderType == "SELL" && triggerPrice != nil && *triggerPrice < price) {
			if *triggerPrice > price {
				return nil, errors.New("Trigger price cannot be greater than price for SL-LIMIT BUY order")
			} else {
				return nil, errors.New("Trigger price cannot be lesser than price for SL-LIMIT SELL order")
			}
		}
	}
	if quantity == 0 {
		return nil, errors.New("Quantity cannot be 0")
	}
	if validity == "" {
		validity = "DAY"
	}

	// Prepare JSON parameters
	jsonParams := map[string]interface{}{
		"exchange":      exchange,
		"order_type":    orderType,
		"price":         price,
		"price_type":    priceType,
		"product_type":  productType,
		"quantity":      quantity,
		"slices":        slices,
		"tradingsymbol": tradingsymbol,
		"validity":      validity,
	}

	// Optional parameters
	if amo != nil {
		jsonParams["amo"] = *amo
	}
	if bookLossPrice != nil {
		jsonParams["book_loss_price"] = *bookLossPrice
	}
	if bookProfitPrice != nil {
		jsonParams["book_profit_price"] = *bookProfitPrice
	}
	if disclosedQuantity != nil {
		jsonParams["disclosed_quantity"] = *disclosedQuantity
	}
	if marketProtection != nil {
		jsonParams["market_protection"] = *marketProtection
	}
	if remarks != nil {
		jsonParams["remarks"] = *remarks
	}
	if trailingPrice != nil {
		jsonParams["trailing_price"] = *trailingPrice
	}
	if triggerPrice != nil {
		jsonParams["trigger_price"] = *triggerPrice
	}

	// Send POST request to slice order
	response, err := c.c2i.sendRequest(c.c2i.BaseURL, "sliceorder", "POST", nil, jsonParams, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Helper functions for validation
// func isValidExchange(exchange string) bool {
// 	// Define valid exchanges here
// 	validExchanges := map[string]bool{"NSE": true, "BSE": true, "NFO": true, "CDS": true, "MCX": true}
// 	return validExchanges[exchange]
// }

// func isValidOrderType(orderType string) bool {
// 	validOrderTypes := map[string]bool{"BUY": true, "SELL": true}
// 	return validOrderTypes[orderType]
// }

// func isValidPriceType(priceType string) bool {
// 	validPriceTypes := map[string]bool{"MARKET": true, "LIMIT": true, "SL-MARKET": true, "SL-LIMIT": true}
// 	return validPriceTypes[priceType]
// }

// func isValidProductType(productType string) bool {
// 	validProductTypes := map[string]bool{"CNC": true, "INTRADAY": true, "NORMAL": true}
// 	return validProductTypes[productType]
// }

// ConvertPositionProductType converts an open position's product type.

func (c *IntegrateOrders) ConvertPositionProductType(
	exchange string,
	orderType string,
	previousProduct string,
	productType string,
	quantity int,
	tradingSymbol string,
	positionType string,
) (map[string]interface{}, error) {

	// Validate parameters
	if !contains(c.c2i.ExchangeTypes, exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !contains(c.c2i.OrderTypes, orderType) {
		return nil, errors.New("invalid order type")
	}
	if !contains(c.c2i.ProductTypes, productType) || !contains(c.c2i.ProductTypes, previousProduct) {
		return nil, errors.New("invalid product type")
	}
	if quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}
	if positionType == "" {
		positionType = "DAY"
	}

	// Prepare JSON parameters
	jsonParams := map[string]interface{}{
		"exchange":         exchange,
		"order_type":       orderType,
		"previous_product": previousProduct,
		"product_type":     productType,
		"quantity":         quantity,
		"tradingsymbol":    tradingSymbol,
		"position_type":    positionType,
	}

	// Send request
	response, err := c.c2i.sendRequest(
		c.c2i.BaseURL,
		"productconversion",
		"POST",
		nil,
		jsonParams,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// PlaceGTTOrder places a GTT order.
func (c *IntegrateOrders) PlaceGTTOrder(
	exchange string,
	orderType string,
	price float64,
	quantity int,
	tradingSymbol string,
	alertPrice float64,
	condition string,
) (map[string]interface{}, error) {

	// Validate parameters
	// if !contains(c.c2i.exchangeTypes, exchange) {
	// 	return nil, errors.New("invalid exchange type")
	// }
	// if !contains(c.c2i.orderTypes, orderType) {
	// 	return nil, errors.New("invalid order type")
	// }
	if quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}
	if !contains(c.c2i.GTTConditionTypes, condition) {
		return nil, errors.New("invalid GTT condition")
	}
	if !c.isValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !c.isValidOrderType(orderType) {
		return nil, errors.New("invalid order type")
	}

	// Prepare JSON parameters
	jsonParams := map[string]interface{}{
		"exchange":      exchange,
		"order_type":    orderType,
		"price":         price,
		"quantity":      quantity,
		"tradingsymbol": tradingSymbol,
		"alert_price":   alertPrice,
		"condition":     condition,
	}

	// Send request
	response, err := c.c2i.sendRequest(
		c.c2i.BaseURL,
		"gttplaceorder",
		"POST",
		nil,
		jsonParams,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// type Orders struct {
// 	c2i *types.C2I
// }

func (o *IntegrateOrders) ModifyGTTOrder(
	exchange string,
	alertID string,
	orderType string,
	tradingsymbol string,
	condition string,
	price float64,
	alertPrice float64,
	quantity int,
) (map[string]interface{}, error) {
	// Validate input parameters
	if !o.isValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !o.isValidOrderType(orderType) {
		return nil, errors.New("invalid order type")
	}
	if quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}

	// Prepare JSON parameters for request
	jsonParams := map[string]interface{}{
		"exchange":      exchange,
		"alert_id":      alertID,
		"order_type":    orderType,
		"price":         price,
		"quantity":      quantity,
		"tradingsymbol": tradingsymbol,
		"alert_price":   alertPrice,
		"condition":     condition,
	}

	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"gttmodify",
		"POST",
		nil,
		jsonParams,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) CancelGTTOrder(alertID string) (map[string]interface{}, error) {
	// Prepare URL parameters
	urlParams := map[string]string{
		"alert_id": alertID,
	}

	gttCancel := fmt.Sprintf("gttcancel/%s", alertID)

	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		gttCancel,
		"GET",
		urlParams,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) PlaceOCOOrder(
	exchange string,
	orderType string,
	tradingsymbol string,
	stoplossQuantity int,
	stoplossPrice float64,
	targetQuantity int,
	targetPrice float64,
	remarks *string,
) (map[string]interface{}, error) {
	// Validate input parameters
	if !o.isValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !o.isValidOrderType(orderType) {
		return nil, errors.New("invalid order type")
	}
	if stoplossQuantity == 0 {
		return nil, errors.New("stoploss quantity cannot be 0")
	}
	if targetQuantity == 0 {
		return nil, errors.New("target quantity cannot be 0")
	}

	// Prepare JSON parameters for request
	jsonParams := map[string]interface{}{
		"exchange":          exchange,
		"order_type":        orderType,
		"tradingsymbol":     tradingsymbol,
		"stoploss_quantity": stoplossQuantity,
		"stoploss_price":    stoplossPrice,
		"target_quantity":   targetQuantity,
		"target_price":      targetPrice,
	}
	if remarks != nil {
		jsonParams["remarks"] = *remarks
	}

	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"ocoplaceorder",
		"POST",
		nil,
		jsonParams,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) ModifyOCOOrder(
	exchange string,
	alertID string,
	orderType string,
	tradingsymbol string,
	stoplossQuantity int,
	stoplossPrice float64,
	targetQuantity int,
	targetPrice float64,
	remarks *string,
) (map[string]interface{}, error) {
	// Validate input parameters
	if !o.isValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !o.isValidOrderType(orderType) {
		return nil, errors.New("invalid order type")
	}
	if stoplossQuantity == 0 {
		return nil, errors.New("stoploss quantity cannot be 0")
	}
	if targetQuantity == 0 {
		return nil, errors.New("target quantity cannot be 0")
	}

	// Prepare JSON parameters for request
	jsonParams := map[string]interface{}{
		"exchange":          exchange,
		"alert_id":          alertID,
		"order_type":        orderType,
		"tradingsymbol":     tradingsymbol,
		"stoploss_quantity": stoplossQuantity,
		"stoploss_price":    stoplossPrice,
		"target_quantity":   targetQuantity,
		"target_price":      targetPrice,
	}
	if remarks != nil {
		jsonParams["remarks"] = *remarks
	}

	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"ocomodify",
		"POST",
		nil,
		jsonParams,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) CancelOCOOrder(alertID string) (map[string]interface{}, error) {
	// Prepare URL parameters
	urlParams := map[string]string{
		"alert_id": alertID,
	}
	ocoCancel := fmt.Sprintf("ococancel/%s", alertID)

	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		ocoCancel,
		"GET",
		urlParams,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) Orders() (map[string]interface{}, error) {
	// Retrieve list of orders
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"orders",
		"GET",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) Order(orderID string) (map[string]interface{}, error) {
	// Retrieve status of a specific order
	urlParams := map[string]string{
		"order_id": orderID,
	}
	orderRoute := fmt.Sprintf("order/%s", orderID)
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		orderRoute,
		"GET",
		urlParams,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) GTTOrders() (map[string]interface{}, error) {
	// Retrieve list of GTT orders
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"gttorders",
		"GET",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) Trades() (map[string]interface{}, error) {
	// Retrieve list of trades
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"trades",
		"GET",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) Positions() (map[string]interface{}, error) {
	// Retrieve list of positions
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"positions",
		"GET",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) Holdings() (map[string]interface{}, error) {
	// Retrieve list of holdings
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"holdings",
		"GET",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) Limits() (map[string]interface{}, error) {
	// Retrieve account balance and cash margin details for all segments
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"limits",
		"GET",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) Margins(orders []map[string]interface{}) (map[string]interface{}, error) {
	// Get margin for a list of orders
	jsonParams := map[string]interface{}{
		"basketlists": orders,
	}
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"margins",
		"POST",
		nil,
		jsonParams,
		nil,
		nil,
		nil,
	)
}

func (o *IntegrateOrders) SpanCalculator(positions []map[string]interface{}) (map[string]interface{}, error) {
	// Get span information for a list of positions
	jsonParams := map[string]interface{}{
		"positions": positions,
	}
	return o.c2i.sendRequest(
		o.c2i.BaseURL,
		"spancalculator",
		"POST",
		nil,
		jsonParams,
		nil,
		nil,
		nil,
	)
}
