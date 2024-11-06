package structs

import (
	"net/http"
	"time"
)

type ConnectToIntegrate struct {
	LoginURL               string
	BaseURL                string
	Timeout                time.Duration
	Logging                bool
	Proxies                map[string]string
	UID                    string
	ActID                  string
	APISessionKey          string
	WSSessionKey           string
	HTTPClient             *http.Client
	ExchangeTypes          []string
	OrderTypes             []string
	PriceTypes             []string
	ProductTypes           []string
	SubscriptionTypes      []string
	ReqSess                *http.Client
	SessionExpiredCallback func()
	GTTConditionTypes      []string
	TimeframeTypes         []string
	Symbols                chan map[string]interface{}
	// SuserToken             string
}

type ModifyOrderParams struct {
	Exchange          string
	OrderID           string
	OrderType         string
	Price             float64
	PriceType         string
	ProductType       string
	Quantity          int
	TradingSymbol     string
	Amo               *string  // Optional
	BookLossPrice     *float64 // Optional
	BookProfitPrice   *float64 // Optional
	DisclosedQuantity *int     // Optional
	MarketProtection  *float64 // Optional
	Remarks           *string  // Optional
	TrailingPrice     *float64 // Optional
	TriggerPrice      *float64 // Optional
	Validity          string
}
