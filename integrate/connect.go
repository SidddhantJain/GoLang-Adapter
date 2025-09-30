package integrate

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"pyintegrate/structs"
	"strconv"
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
		Logging:           logging,
		Timeout:           time.Duration(timeout) * time.Second,
		Proxies:           proxies,
		ReqSess:           &http.Client{Timeout: time.Duration(timeout) * time.Second},
		LoginURL:          loginURL,
		BaseURL:           baseURL,
		Symbol:            []structs.Symbol{},
		ExchangeTypes:     []string{"NSE", "BSE", "NFO", "CDS", "MCX"},
		OrderTypes:        []string{"BUY", "SELL"},
		PriceTypes:        []string{"MARKET", "LIMIT", "SL-MARKET", "SL-LIMIT"},
		ProductTypes:      []string{"CNC", "INTRADAY", "NORMAL"},
		SubscriptionTypes: []string{"TICK", "ORDER", "DEPTH"},
		GTTConditionTypes: []string{"LTP_ABOVE", "LTP_BELOW"},
		TimeframeTypes:    []string{"minute", "day", "tick"},
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
	// Fetch and store symbols
	if err := Symbols(c); err != nil {
		return fmt.Errorf("failed to fetch symbols: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

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
	routePrefix string, route string, method string,
	urlParams map[string]interface{},
	jsonParams map[string]interface{},
	dataParams map[string]interface{},
	queryParams map[string]interface{},
	extraHeaders map[string]interface{},
) (map[string]interface{}, error) {
	// Build URL
	// fullURL := routePrefix + route
	fullURL := routePrefix
	if urlParams != nil {
		routeTmpl := route
		for k, v := range urlParams {
			routeTmpl = strings.ReplaceAll(routeTmpl, "{"+k+"}", fmt.Sprintf("%v", v))
		}
		fullURL = strings.TrimRight(routePrefix, "/") + "/" + strings.TrimLeft(routeTmpl, "/")
	} else {
		fullURL = strings.TrimRight(routePrefix, "/") + "/" + strings.TrimLeft(route, "/")
	}
	if queryParams != nil {
		//query := url.Values{}
		//for k, v := range queryParams {
		//	query.Add(k, fmt.Sprintf("%v", v))
		//}
		//fullURL += "?" + query.Encode()
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()

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
	//client := &http.Client{Timeout: s.Timeout}
	resp, err := s.ReqSess.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received non-2xx response: %d %s", resp.StatusCode, resp.Status)
	}
	var data map[string]interface{}
	// Parse Response
	contentType := resp.Header.Get("content-type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}
	} else if strings.HasPrefix(contentType, "text/csv") {
		csvBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV response: %w", err)
		}
		data = map[string]interface{}{
			"data": string(csvBytes),
		}
	} else {
		// Log and return raw body for debugging
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Unexpected Content-Type: %s\nRaw response: %s\n", contentType, string(bodyBytes))
		return nil, fmt.Errorf("unexpected content-type: %s", contentType)
	}
	return data, nil
}

func Symbols(s *LocalConnect) error {
	symbolFileName := filepath.Join(filepath.Dir(os.Args[0]), "allmaster.csv")
	fileInfo, err := os.Stat(symbolFileName)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		//route := "https://app.definedgesecurities.com/public/allmaster.zip"
		req, err := http.NewRequest(
			"GET",
			"https://app.definedgesecurities.com/public/allmaster.zip",
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := s.ReqSess.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logger.Println("Failed to close response body:", err)
				fmt.Println("Failed to close response body :", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		zipBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read zip file: %w", err)
		}
		zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
		if err != nil {
			return fmt.Errorf("failed to open zip file: %w", err)
		}
		found := false
		for _, f := range zipReader.File {
			if f.Name == "allmaster.csv" {
				rc, err := f.Open()
				if err != nil {
					return fmt.Errorf("failed to open csv inside zip: %w", err)
				}
				defer func(rc io.ReadCloser) {
					err := rc.Close()
					if err != nil {
						fmt.Println("Failed to close csv file:", err)
						logger.Println("Failed to close csv file:", err)
					}
				}(rc)

				out, err := os.Create(symbolFileName)
				if err != nil {
					return fmt.Errorf("failed to create symbol file: %w", err)
				}
				defer func(out *os.File) {
					err := out.Close()
					if err != nil {
						fmt.Println("Failed to close file:", err)
						logger.Println("Failed to close file:", err)
					}
				}(out)

				_, err = io.Copy(out, rc)
				if err != nil {
					return fmt.Errorf("failed to extract csv: %w", err)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("allmaster.csv not found in zip")
		}

		// After extraction, parse the CSV and load symbols
		file, err := os.Open(symbolFileName)
		if err != nil {
			return fmt.Errorf("failed to open symbol file: %w", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println("Failed to close symbol file:", err)
				logger.Println("Failed to close symbol file:", err)
			}
		}(file)

		s.Symbol = nil // clear previous symbols
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read symbol file: %w", err)
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			cols := strings.Split(line, ",")
			if len(cols) < 14 {
				continue // skip incomplete lines
			}
			strike := ""
			if len(cols) > 11 && len(cols) > 10 && len(cols) > 9 {
				// strike = str(int(int(line[9]) / (int(line[11]) * 10 ** int(line[10]))))
				strikeInt := 0
				base, err1 := strconv.Atoi(cols[9])
				mult, err2 := strconv.Atoi(cols[11])
				pow, err3 := strconv.Atoi(cols[10])
				if err1 == nil && err2 == nil && err3 == nil && mult != 0 {
					strikeInt = base / (mult * int(math.Pow10(pow)))
					strike = strconv.Itoa(strikeInt)
				}
			}
			s.Symbol = append(s.Symbol, structs.Symbol{
				Segment:        cols[0],
				Token:          cols[1],
				Symbol:         cols[2],
				TradingSymbol:  cols[3],
				InstrumentType: cols[4],
				Expiry:         cols[5],
				TickSize:       cols[6],
				LotSize:        cols[7],
				OptionType:     cols[8],
				Strike:         strike,
				ISIN:           cols[12],
				PriceMult:      cols[13],
			})
		}
		//}
	}
	return nil
}
