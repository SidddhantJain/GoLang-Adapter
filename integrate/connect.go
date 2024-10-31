package integrate

import (
    "archive/zip"
    "bytes"
    "crypto/sha256"
    "encoding/csv"
    "encoding/hex"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "time"
)

type ConnectToIntegrate struct {
    LoginURL             string
    BaseURL              string
    Timeout              time.Duration
    Logging              bool
    Proxies              map[string]string
    Uid                  string
    Actid                string
    APISessionKey        string
    WSSessionKey         string
    HTTPClient           *http.Client
    ExchangeTypes        []string
    OrderTypes           []string
    PriceTypes           []string
    ProductTypes         []string
    SubscriptionTypes    []string
}

// Set up a logger
var logger = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile)

// Constants for exchanges
const (
	ExchangeTypeNSE = "NSE"
	ExchangeTypeBSE = "BSE"
	ExchangeTypeNFO = "NFO"
	ExchangeTypeCDS = "CDS"
	ExchangeTypeMCX = "MCX"
)

// Constants for order types
const (
	OrderTypeBuy  = "BUY"
	OrderTypeSell = "SELL"
)

// Constants for price types
const (
	PriceTypeMarket  = "MARKET"
	PriceTypeLimit   = "LIMIT"
	PriceTypeSlMkt   = "SL-MARKET"
	PriceTypeSlLmt   = "SL-LIMIT"
)

// Constants for product types
const (
	ProductTypeCNC      = "CNC"
	ProductTypeIntraday = "INTRADAY"
	ProductTypeNormal   = "NORMAL"
)

// Constants for subscription types
const (
	SubscriptionTypeTick  = "TICK"
	SubscriptionTypeOrder = "ORDER"
	SubscriptionTypeDepth = "DEPTH"
)

// Constants for validity types
const (
	ValidityTypeDay = "DAY"
	ValidityTypeIOC = "IOC"
	ValidityTypeEOS = "EOS"
)

// Constants for order statuses
const (
	OrderStatusNew      = "NEW"
	OrderStatusOpen     = "OPEN"
	OrderStatusComplete = "COMPLETE"
	OrderStatusCancelled = "CANCELED"
	OrderStatusRejected  = "REJECTED"
	OrderStatusReplaced  = "REPLACED"
)

// Constants for GTT conditions
const (
	GttConditionLtpBelow = "LTP_BELOW"
	GttConditionLtpAbove = "LTP_ABOVE"
)

// Constants for timeframe types
const (
	TimeframeTypeMin  = "minute"
	TimeframeTypeDay  = "day"
	TimeframeTypeTick = "tick"
)


func NewConnectToIntegrate(loginURL, baseURL string, timeout int, logging bool, proxies map[string]string) *ConnectToIntegrate {
	// Set default URLs if not provided
	if loginURL == "" {
		loginURL = "https://signin.definedgesecurities.com/auth/realms/debroking/dsbpkc/"
	}
	if baseURL == "" {
		baseURL = "https://integrate.definedgesecurities.com/dart/v1/"
	}

	// Default timeout to 10 seconds if not set
	if timeout == 0 {
		timeout = 10
	}

	// Initialize and configure the ConnectToIntegrate instance
	connect := &ConnectToIntegrate{
		Logging:                logging,
		Timeout:                time.Duration(timeout) * time.Second,
		Proxies:                proxies,
		ReqSess:                &http.Client{Timeout: time.Duration(timeout) * time.Second},
		UID:                    "",
		ActID:                  "",
		APISessionKey:          "",
		WSSessionKey:           "",
		LoginURL:               loginURL,
		BaseURL:                baseURL,
		SessionExpiredCallback: nil, // Set a callback function if needed

		// Initialize exchange, order, price, product, and subscription types
		ExchangeTypes:       []string{"NSE", "BSE", "NFO", "CDS", "MCX"},
		OrderTypes:          []string{"BUY", "SELL"},
		PriceTypes:          []string{"MARKET", "LIMIT", "SL-MARKET", "SL-LIMIT"},
		ProductTypes:        []string{"CNC", "INTRADAY", "NORMAL"},
		SubscriptionTypes:   []string{"TICK", "ORDER", "DEPTH"},
		GTTConditionTypes:   []string{"LTP_BELOW", "LTP_ABOVE"},
		TimeframeTypes:      []string{"minute", "day", "tick"},
	}
  }


//Login 
func (c *ConnectToIntegrate) login(apiToken, apiSecret string, totp *string) error {
	if apiToken == "" || apiSecret == "" {
		return errors.New("invalid api_token or api_secret")
	}

	// Get OTP token
	r, err := c.sendRequest(c.loginURL, "login/"+apiToken, "GET", map[string]string{"api_secret": apiSecret}, nil)
	if err != nil {
		return err
	}

	otpToken, ok := r["otp_token"].(string)
	if !ok {
		return errors.New("failed to obtain otp_token")
	}

	// Get OTP/TOTP for 2FA
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

	// Compute the session key
	ac := sha256.New()
	ac.Write([]byte(otpToken + otp + apiSecret))
	acHex := hex.EncodeToString(ac.Sum(nil))

	// Get session keys
	r, err = c.sendRequest(c.loginURL, "token", "POST", nil, map[string]interface{}{
		"otp_token": otpToken,
		"otp":       otp,
		"ac":        acHex,
	})
	if err != nil {
		return err
	}

	// Set session keys
	c.setSessionKeys(r["uid"].(string), r["actid"].(string), r["api_session_key"].(string))

	// Attempt to remove symbols file
	symbolsFilename := filepath.Join(filepath.Dir(os.Args[0]), "allmaster.csv")
	if err := os.Remove(symbolsFilename); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Call next on the symbols channel
	select {
	case c.symbols <- struct{}{}:
	default:
	}

	return nil
}



// getSessionKeys retrieves stored session keys
// Returns the session keys as strings.
func (c *ConnectToIntegrate) getSessionKeys() (string, string, string, string) {
	return c.uid, c.actid, c.apiSessionKey, c.wsSessionKey
}

// setSessionKeys stores session keys
//
// Parameters:
//   uid: Your Definedge Securities login UCC id
//   actid: Your Definedge Securities login account id
//   apiSessionKey: Your Definedge Securities API session key
//   wsSessionKey: Your Definedge Securities WebSocket session key
func (c *ConnectToIntegrate) setSessionKeys(uid, actid, apiSessionKey, wsSessionKey string) {
	c.uid = uid
	c.actid = actid
	c.apiSessionKey = apiSessionKey
	c.wsSessionKey = wsSessionKey
}


// SymbolsGenerator returns a channel that yields symbols
func SymbolsGenerator() <-chan Symbol {
	symbolsChannel := make(chan Symbol)

	go func() {
		defer close(symbolsChannel)

		// Path for the symbols file
		symbolsFilename := filepath.Join("allmaster.csv")

		// Check if the file exists
		if _, err := os.Stat(symbolsFilename); os.IsNotExist(err) {
			// Download the master file if not present
			err := downloadSymbols()
			if err != nil {
				fmt.Println("Error downloading symbols:", err)
				return
			}
		}

		// Open the symbols file
		file, err := os.Open(symbolsFilename)
		if err != nil {
			fmt.Println("Error opening symbols file:", err)
			return
		}
		defer file.Close()

		// Read the CSV file
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			fmt.Println("Error reading CSV file:", err)
			return
		}

		// Create and yield symbols
		for _, record := range records {
			if len(record) < 14 { // Ensure there are enough columns
				continue
			}
			symbol := Symbol{
				Segment:        record[0],
				Token:          record[1],
				Symbol:         record[2],
				TradingSymbol:  record[3],
				InstrumentType: record[4],
				Expiry:         record[5],
				TickSize:       record[6],
				LotSize:        record[7],
				OptionType:     record[8],
				Strike:         fmt.Sprintf("%d", int(int(record[9])/ (int(record[11]) * 10 ^ int(record[10])))), // Convert Strike to int and format
				ISIN:           record[12],
				PriceMult:      record[13],
			}
			symbolsChannel <- symbol
		}
	}()

	return symbolsChannel
}

// downloadSymbols downloads the symbols file
func downloadSymbols() error {
	url := "https://app.definedgesecurities.com/public/allmaster.zip"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the zip content
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	// Extract the CSV file from the zip
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return err
	}

	for _, file := range zr.File {
		if file.Name == "allmaster.csv" {
			outFile, err := os.Create("allmaster.csv")
			if err != nil {
				return err
			}
			defer outFile.Close()

			reader, err := file.Open()
			if err != nil {
				return err
			}
			defer reader.Close()

			_, err = io.Copy(outFile, reader)
			return err
		}
	}
	return fmt.Errorf("allmaster.csv not found in zip")
}


//function to send request
func (s *YourStruct) sendRequest(
	routePrefix string,
	route string,
	method string,
	urlParams map[string]string,
	jsonParams map[string]interface{},
	dataParams map[string]interface{},
	queryParams map[string]string,
	extraHeaders map[string]string,
) (map[string]interface{}, error) {
	// Form URL
	urlStr := routePrefix + fmt.Sprintf(route, urlParams)
	if queryParams != nil {
		query := url.Values{}
		for k, v := range queryParams {
			query.Add(k, v)
		}
		urlStr += "?" + query.Encode()
	}

	// Create a new HTTP request
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	headers := make(map[string]string)
	if extraHeaders != nil {
		for k, v := range extraHeaders {
			headers[k] = v
		}
	}
	if s.APIKey != "" {
		headers["Authorization"] = s.APIKey
	}

	// Add headers to request
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Set request body based on method
	if method == http.MethodPost {
		if jsonParams != nil {
			jsonData, err := json.Marshal(jsonParams)
			if err != nil {
				return nil, err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
		} else if dataParams != nil {
			formData, err := json.Marshal(dataParams)
			if err != nil {
				return nil, err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(formData))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// Logging the request
	if s.Logging {
		fmt.Printf("Request: %s %s %v\n", method, urlStr, headers)
	}

	// Make the HTTP request
	client := &http.Client{
		Timeout: s.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log response
	if s.Logging {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response: %d %s\n", resp.StatusCode, bodyBytes)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body
	}

	// Check Content-Type and handle the response
	var data map[string]interface{}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("Couldn't parse JSON response: %s", err)
		}
	} else if contentType == "text/csv" {
		csvReader := csv.NewReader(resp.Body)
		records, err := csvReader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse CSV response: %s", err)
		}
		data = map[string]interface{}{"data": records}
	} else {
		return nil, fmt.Errorf("Unknown Content-Type (%s): %s", contentType, resp.Status)
	}

	// Handle response status
	if status, exists := data["status"]; exists {
		if status == "ERROR" {
			if s.SessionExpiredCallback != nil && data["message"] == "Session Expired" {
				s.SessionExpiredCallback()
				if s.Logging {
					fmt.Println("Session expired. Callback called")
				}
			} else {
				return nil, fmt.Errorf("Error: %v", data)
			}
		} else if status == "SUCCESS" && resp.Request.URL.String() == fmt.Sprintf("%s/sliceorder", s.BaseURL) {
			if orders, ok := data["orders"].([]interface{}); ok {
				for _, order := range orders {
					if orderMap, ok := order.(map[string]interface{}); ok && orderMap["status"] == "ERROR" {
						return nil, fmt.Errorf("Error: %v", data)
					}
				}
			}
		}
	}

	return data, nil
}
