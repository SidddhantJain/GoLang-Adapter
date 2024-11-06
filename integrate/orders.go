package integrate

import (
	"adapter-project/structs"
	"errors"
	"fmt"
	"net/http"
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
	if !contains(c.c2i.ExchangeTypes, params.Exchange) {
		return nil, errors.New("invalid exchange type")
	}

	// Check order type
	if !contains(c.OrderTypes, params.OrderType) {
		return nil, errors.New("invalid order type")
	}

	// Check price type
	if !contains(c.PriceTypes, params.PriceType) {
		return nil, errors.New("invalid price type")
	}

	// Check product type
	if !contains(c.ProductTypes, params.ProductType) {
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
	response, err := c.SendRequest("modify", http.MethodPost, jsonParams)
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
func (c *Client) CancelOrder(orderID string) (map[string]interface{}, error) {
	if orderID == "" {
		return nil, errors.New("order ID cannot be empty")
	}

	// Prepare the route for cancellation
	route := fmt.Sprintf("cancel/%s", orderID)

	// Send GET request for cancellation
	response, err := c.sendRequest("GET", c.baseURL, route, nil, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// SliceOrder slices an order into multiple parts and places each as a separate order.
func (c *Client) SliceOrder(
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
	if !isValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !isValidOrderType(orderType) {
		return nil, errors.New("invalid order type")
	}
	if !isValidPriceType(priceType) {
		return nil, errors.New("invalid price type")
	}
	if !isValidProductType(productType) {
		return nil, errors.New("invalid product type")
	}
	if priceType == "MARKET" && price != 0 {
		return nil, errors.New("price should be 0 for market order")
	}
	if priceType == "SL-LIMIT" {
		if (orderType == "BUY" && triggerPrice != nil && *triggerPrice > price) ||
			(orderType == "SELL" && triggerPrice != nil && *triggerPrice < price) {
			return nil, errors.New("invalid trigger price for SL-LIMIT order")
		}
	}
	if quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}

	// Prepare JSON parameters
	jsonParams := map[string]interface{}{
		"exchange":      exchange,
		"orderType":     orderType,
		"price":         price,
		"priceType":     priceType,
		"productType":   productType,
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
		jsonParams["bookLossPrice"] = *bookLossPrice
	}
	if bookProfitPrice != nil {
		jsonParams["bookProfitPrice"] = *bookProfitPrice
	}
	if disclosedQuantity != nil {
		jsonParams["disclosedQuantity"] = *disclosedQuantity
	}
	if marketProtection != nil {
		jsonParams["marketProtection"] = *marketProtection
	}
	if remarks != nil {
		jsonParams["remarks"] = *remarks
	}
	if trailingPrice != nil {
		jsonParams["trailingPrice"] = *trailingPrice
	}
	if triggerPrice != nil {
		jsonParams["triggerPrice"] = *triggerPrice
	}

	// Send POST request to slice order
	response, err := c.sendRequest("POST", c.baseURL, "sliceorder", jsonParams, nil)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Helper functions for validation
func isValidExchange(exchange string) bool {
	// Define valid exchanges here
	validExchanges := map[string]bool{"NSE": true, "BSE": true, "NFO": true, "CDS": true, "MCX": true}
	return validExchanges[exchange]
}

func isValidOrderType(orderType string) bool {
	validOrderTypes := map[string]bool{"BUY": true, "SELL": true}
	return validOrderTypes[orderType]
}

func isValidPriceType(priceType string) bool {
	validPriceTypes := map[string]bool{"MARKET": true, "LIMIT": true, "SL-MARKET": true, "SL-LIMIT": true}
	return validPriceTypes[priceType]
}

func isValidProductType(productType string) bool {
	validProductTypes := map[string]bool{"CNC": true, "INTRADAY": true, "NORMAL": true}
	return validProductTypes[productType]
}

// ConvertPositionProductType converts an open position's product type.
func (c *Client) ConvertPositionProductType(
	exchange string,
	orderType string,
	previousProduct string,
	productType string,
	quantity int,
	tradingSymbol string,
	positionType string,
) (map[string]interface{}, error) {

	// Validate parameters
	if !contains(c.c2i.exchangeTypes, exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !contains(c.c2i.orderTypes, orderType) {
		return nil, errors.New("invalid order type")
	}
	if !contains(c.c2i.productTypes, productType) || !contains(c.c2i.productTypes, previousProduct) {
		return nil, errors.New("invalid product type")
	}
	if quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}

	// Prepare JSON parameters
	jsonParams := map[string]interface{}{
		"exchange":        exchange,
		"orderType":       orderType,
		"previousProduct": previousProduct,
		"productType":     productType,
		"quantity":        quantity,
		"tradingSymbol":   tradingSymbol,
		"positionType":    positionType,
	}

	// Send request
	response, err := c.c2i.SendRequest(
		c.c2i.BaseURL,
		"productconversion",
		"POST",
		jsonParams,
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// PlaceGTTOrder places a GTT order.
func (c *Client) PlaceGTTOrder(
	exchange string,
	orderType string,
	price float64,
	quantity int,
	tradingSymbol string,
	alertPrice float64,
	condition string,
) (map[string]interface{}, error) {

	// Validate parameters
	if !contains(c.c2i.exchangeTypes, exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !contains(c.c2i.orderTypes, orderType) {
		return nil, errors.New("invalid order type")
	}
	if quantity == 0 {
		return nil, errors.New("quantity cannot be 0")
	}
	if !contains(c.c2i.gttConditionTypes, condition) {
		return nil, errors.New("invalid GTT condition")
	}

	// Prepare JSON parameters
	jsonParams := map[string]interface{}{
		"exchange":      exchange,
		"orderType":     orderType,
		"price":         price,
		"quantity":      quantity,
		"tradingSymbol": tradingSymbol,
		"alertPrice":    alertPrice,
		"condition":     condition,
	}

	// Send request
	response, err := c.c2i.SendRequest(
		c.c2i.BaseURL,
		"gttplaceorder",
		"POST",
		jsonParams,
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Utility function to check if a string exists in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type Orders struct {
	c2i *types.C2I
}

func (o *Orders) ModifyGTTOrder(
	exchange, alertID, orderType, tradingsymbol, condition string,
	price, alertPrice float64,
	quantity int,
) (map[string]interface{}, error) {
	// Validate input parameters
	if !o.c2i.IsValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !o.c2i.IsValidOrderType(orderType) {
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

	return o.c2i.SendRequest("gttmodify", "POST", jsonParams)
}

func (o *Orders) CancelGTTOrder(alertID string) (map[string]interface{}, error) {
	// Prepare URL parameters
	urlParams := map[string]string{
		"alert_id": alertID,
	}

	return o.c2i.SendRequestWithURLParams("gttcancel/{alert_id}", "GET", nil, urlParams)
}

func (o *Orders) PlaceOCOOrder(
	exchange, orderType, tradingsymbol string,
	stoplossQuantity int, stoplossPrice float64,
	targetQuantity int, targetPrice float64,
	remarks *string,
) (map[string]interface{}, error) {
	// Validate input parameters
	if !o.c2i.IsValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !o.c2i.IsValidOrderType(orderType) {
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

	return o.c2i.SendRequest("ocoplaceorder", "POST", jsonParams)
}

func (o *Orders) ModifyOCOOrder(
	exchange, alertID, orderType, tradingsymbol string,
	stoplossQuantity int, stoplossPrice float64,
	targetQuantity int, targetPrice float64,
	remarks *string,
) (map[string]interface{}, error) {
	// Validate input parameters
	if !o.c2i.IsValidExchange(exchange) {
		return nil, errors.New("invalid exchange type")
	}
	if !o.c2i.IsValidOrderType(orderType) {
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

	return o.c2i.SendRequest("ocomodify", "POST", jsonParams)
}

func (o *Orders) CancelOCOOrder(alertID string) (map[string]interface{}, error) {
	// Prepare URL parameters
	urlParams := map[string]string{
		"alert_id": alertID,
	}

	return o.c2i.SendRequestWithURLParams("ococancel/{alert_id}", "GET", nil, urlParams)
}

func (o *Orders) Orders() (map[string]interface{}, error) {
	// Retrieve list of orders
	return o.c2i.SendRequest("orders", "GET", nil)
}

func (o *Orders) Order(orderID string) (map[string]interface{}, error) {
	// Retrieve status of a specific order
	urlParams := map[string]string{
		"order_id": orderID,
	}
	return o.c2i.SendRequestWithURLParams("order/{order_id}", "GET", nil, urlParams)
}

func (o *Orders) GTTOrders() (map[string]interface{}, error) {
	// Retrieve list of GTT orders
	return o.c2i.SendRequest("gttorders", "GET", nil)
}

func (o *Orders) Trades() (map[string]interface{}, error) {
	// Retrieve list of trades
	return o.c2i.SendRequest("trades", "GET", nil)
}

func (o *Orders) Positions() (map[string]interface{}, error) {
	// Retrieve list of positions
	return o.c2i.SendRequest("positions", "GET", nil)
}

func (o *Orders) Holdings() (map[string]interface{}, error) {
	// Retrieve list of holdings
	return o.c2i.SendRequest("holdings", "GET", nil)
}

func (o *Orders) Limits() (map[string]interface{}, error) {
	// Retrieve account balance and cash margin details for all segments
	return o.c2i.SendRequest("limits", "GET", nil)
}

func (o *Orders) Margins(orders []map[string]interface{}) (map[string]interface{}, error) {
	// Get margin for a list of orders
	jsonParams := map[string]interface{}{
		"basketlists": orders,
	}
	return o.c2i.SendRequest("margin", "POST", jsonParams)
}

func (o *Orders) SpanCalculator(positions []map[string]interface{}) (map[string]interface{}, error) {
	// Get span information for a list of positions
	jsonParams := map[string]interface{}{
		"positions": positions,
	}
	return o.c2i.SendRequest("spancalculator", "POST", jsonParams)
}
