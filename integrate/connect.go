package integrate

import (
	"adapter-project/structs"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalConnect provides an interface to interact with Definedge Securities API.
type LocalConnect struct {
	*structs.ConnectToIntegrate
}

var logger = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile)

// Constants for exchanges, orders, prices, etc.
const (
	ExchangeTypeNSE       = "NSE"
	ExchangeTypeBSE       = "BSE"
	ExchangeTypeNFO       = "NFO"
	ExchangeTypeCDS       = "CDS"
	ExchangeTypeMCX       = "MCX"
	OrderTypeBuy          = "BUY"
	OrderTypeSell         = "SELL"
	PriceTypeMarket       = "MARKET"
	PriceTypeLimit        = "LIMIT"
	PriceTypeSlMkt        = "SL-MARKET"
	PriceTypeSlLmt        = "SL-LIMIT"
	ProductTypeCNC        = "CNC"
	ProductTypeIntraday   = "INTRADAY"
	ProductTypeNormal     = "NORMAL"
	SubscriptionTypeTick  = "TICK"
	SubscriptionTypeOrder = "ORDER"
	SubscriptionTypeDepth = "DEPTH"
	ValidityTypeDay       = "DAY"
	ValidityTypeIOC       = "IOC"
	ValidityTypeEOS       = "EOS"
	OrderStatusNew        = "NEW"
	OrderStatusOpen       = "OPEN"
	OrderStatusComplete   = "COMPLETE"
	OrderStatusCancelled  = "CANCELED"
	OrderStatusRejected   = "REJECTED"
	OrderStatusReplaced   = "REPLACED"
	GttConditionLtpAbove  = "LTP_ABOVE"
	GttConditionLtpBelow  = "LTP_BELOW"
	TimeframeTypeMin      = "minute"
	TimeframeTypeDay      = "day"
	TimeframeTypeTick     = "tick"
)

// NewConnectToIntegrate initializes the API client.
func NewConnectToIntegrate(
	loginURL string,
	baseURL string,
	timeout int,
	logging bool,
	proxies map[string]string,
) *LocalConnect {
	if loginURL == "" {
		loginURL = "https://signin.definedgebroking.com/auth/realms/debroking/dsbpkc/"
	}
	if baseURL == "" {
		baseURL = "https://api.definedgebroking.com/dart/v1/"
	}
	if timeout == 0 {
		timeout = 10
	}
	connect := &structs.ConnectToIntegrate{
		Logging:  logging,
		Timeout:  time.Duration(timeout) * time.Second,
		Proxies:  proxies,
		ReqSess:  &http.Client{Timeout: time.Duration(timeout) * time.Second},
		LoginURL: loginURL,
		BaseURL:  baseURL,
		Symbols:  make(chan map[string]interface{}),
	}
	return &LocalConnect{connect}
}

// Login authenticates the user with the API.
func (c *LocalConnect) Login(apiToken string, apiSecret string, totp *string) error {
	if apiToken == "" || apiSecret == "" {
		return errors.New("invalid api_token or api_secret")
	}

	headers := map[string]interface{}{"api_secret": apiSecret}
	route := fmt.Sprintf("login/%s", apiToken)

	// Step 1: Get OTP Token
	response, err := c.sendRequest(c.LoginURL, route, "GET", nil, nil, nil, nil, headers)
	if err != nil {
		return fmt.Errorf("failed to get OTP token: %w", err)
	}

	fmt.Printf("Response from OTP request: %+v\n", response) // Debug print

	otpToken, ok := response["otp_token"].(string)
	if !ok || otpToken == "" {
		return errors.New("failed to obtain otp_token")
	}

	// Step 2: Get OTP
	var otp string
	if totp == nil {
		fmt.Print("Enter OTP/External TOTP: ")
		_, err := fmt.Scan(&otp)
		if err != nil {
			return errors.New("no OTP/TOTP provided")
		}
	} else {
		otp = *totp
	}
	// print(otp, totp)
	// Step 3: Generate Session Key
	ac := sha256.New()
	ac.Write([]byte(otpToken + otp + apiSecret))
	acHex := hex.EncodeToString(ac.Sum(nil))

	// Step 4: Obtain Session Keys
	response, err = c.sendRequest(c.LoginURL, "token", "POST", nil, map[string]interface{}{
		"otp_token": otpToken,
		"otp":       otp,
		"ac":        acHex,
	}, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to obtain session keys: %w", err)
	}
	// if uncomment the below code without the responce it gives an error
	// uid, uidOk := response["uid"].(string)
	// actid, actidOk := response["actid"].(string)
	// apiSessionKey, apiSessionKeyOk := response["api_session_key"].(string)
	// wsSessionKey, wsSessionKeyOk := response["susertoken"].(string)
	// if !uidOk || !actidOk || !apiSessionKeyOk || !wsSessionKeyOk {
	// 	return errors.New("missing or invalid keys in API response")
	// }

	// fmt.Printf("uid %s\n", uid)                       // Debug print
	// fmt.Printf("actid %s\n", actid)                   // Debug print
	// fmt.Printf("api_session_key %s\n", apiSessionKey) // Debug print
	// fmt.Printf("susertoken %s\n", wsSessionKey)       // Debug print
	// fmt.Print(response)

	// Store session keys
	// GIVING AN ERROR IN UID
	c.setSessionKeys(
		response["uid"].(string),
		response["actid"].(string),
		response["api_session_key"].(string),
		response["susertoken"].(string),
	)

	// Remove any existing symbols file
	symbolsFile := filepath.Join(filepath.Dir(os.Args[0]), "allmaster.csv")
	if err := os.Remove(symbolsFile); err != nil && !os.IsNotExist(err) {
		logger.Println("Symbols file not found or failed to delete.")
	}
	time.Sleep(3 * time.Millisecond) // its very necessary

	return nil
}

// setSessionKeys stores session keys.
func (c *LocalConnect) setSessionKeys(uid string, actid string, apiSessionKey string, wsSessionKey string) {
	c.UID = uid
	c.ActID = actid
	c.APISessionKey = apiSessionKey
	c.WSSessionKey = wsSessionKey
}

// sendRequest handles API requests and responses.
func (s *LocalConnect) sendRequest(
	routePrefix, route, method string,
	urlParams map[string]interface{},
	jsonParams map[string]interface{},
	dataParams map[string]interface{},
	queryParams map[string]interface{},
	extraHeaders map[string]interface{},
) (map[string]interface{}, error) {
	// Build URL
	fullURL := routePrefix + route
	if queryParams != nil {
		query := url.Values{}
		for k, v := range queryParams {
			query.Add(k, fmt.Sprintf("%v", v))
		}
		fullURL += "?" + query.Encode()
	}

	// Prepare request body
	var body io.Reader
	if jsonParams != nil {
		jsonData, err := json.Marshal(jsonParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON params: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	} else if dataParams != nil {
		formData := url.Values{}
		for k, v := range dataParams {
			formData.Add(k, fmt.Sprintf("%v", v))
		}
		body = strings.NewReader(formData.Encode())
	}

	// Create Request
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Headers
	for k, v := range extraHeaders {
		req.Header.Add(k, fmt.Sprintf("%v", v))
	}
	if s.APISessionKey != "" {
		// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APISessionKey))
		// logger.Printf("Authorization Header: Bearer %s", s.APISessionKey)
		req.Header.Set("Authorization", s.APISessionKey)
	}
	if jsonParams != nil {
		req.Header.Set("Content-Type", "application/json")
	} else if dataParams != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Send Request
	client := &http.Client{Timeout: s.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received non-2xx response: %d %s", resp.StatusCode, resp.Status)
	}

	// Parse Response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return data, nil
}
