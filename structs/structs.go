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
	SessionExpiredCallback func(err error)
	GTTConditionTypes      []string
	TimeframeTypes         []string
	// SuserToken             string
}
