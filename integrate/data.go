package integrate

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
    "time"
)

type IntegrateData struct {
    c2i     *ConnectToIntegrate
    logging bool
}

// NewIntegrateData initializes a new instance of IntegrateData
func NewIntegrateData(c2i *ConnectToIntegrate, logging bool) *IntegrateData {
    return &IntegrateData{
        c2i:     c2i,
        logging: logging,
    }
}

// HistoricalData retrieves historical data for a security.
// Returns data as a channel of maps, similar to Python's generator.
func (ic *IntegrateData) HistoricalData(exchange, tradingSymbol, timeframe string, start, end time.Time) (<-chan map[string]interface{}, <-chan error, error) {
    if !ic.isValidExchange(exchange) {
        return nil, nil, errors.New("invalid exchange type")
    }
    if !ic.isValidTimeframe(timeframe) {
        return nil, nil, errors.New("invalid timeframe")
    }

    token, err := ic.getToken(exchange, tradingSymbol)
    if err != nil {
        return nil, nil, err
    }

    route := fmt.Sprintf("https://data.definedgesecurities.com/sds/history/%s/%s/%s/%s/%s",
        exchange, token, timeframe, start.Format("020120061504"), end.Format("020120061504"))

    dataChan := make(chan map[string]interface{})
    errorChan := make(chan error, 1)

    go func() {
        defer close(dataChan)
        defer close(errorChan)

        response, err := ic.c2i.sendRequest(route, "GET")
        if err != nil {
            errorChan <- err
            return
        }

        data, ok := response["data"].([]interface{})
        if !ok {
            errorChan <- errors.New("unexpected response format")
            return
        }

        for _, line := range data {
            dataStr, ok := line.(string)
            if !ok {
                continue
            }
            fields := parseFields(dataStr)

            if len(fields) == 7 {
                dataChan <- map[string]interface{}{
                    "datetime": parseDate(fields[0]),
                    "open":     toFloat(fields[1]),
                    "high":     toFloat(fields[2]),
                    "low":      toFloat(fields[3]),
                    "close":    toFloat(fields[4]),
                    "volume":   toInt(fields[5]),
                    "oi":       toInt(fields[6]),
                }
            } else if len(fields) == 4 {
                dataChan <- map[string]interface{}{
                    "utc": fields[0],
                    "ltp": toFloat(fields[1]),
                    "ltq": toFloat(fields[2]),
                    "oi":  toFloat(fields[3]),
                }
            }
        }
    }()
    return dataChan, errorChan, nil
}

// Quotes retrieves the quote for a security.
func (ic *IntegrateData) Quotes(exchange, tradingSymbol string) (map[string]interface{}, error) {
    if !ic.isValidExchange(exchange) {
        return nil, errors.New("invalid exchange type")
    }

    token, err := ic.getToken(exchange, tradingSymbol)
    if err != nil {
        return nil, err
    }

    route := fmt.Sprintf("%s/quotes/%s/%s", ic.c2i.baseURL, exchange, token)
    return ic.c2i.sendRequest(route, "GET")
}

// SecurityInformation retrieves information about a security.
func (ic *IntegrateData) SecurityInformation(exchange, tradingSymbol string) (map[string]interface{}, error) {
    if !ic.isValidExchange(exchange) {
        return nil, errors.New("invalid exchange type")
    }

    token, err := ic.getToken(exchange, tradingSymbol)
    if err != nil {
        return nil, err
    }

    route := fmt.Sprintf("%s/securityinfo/%s/%s", ic.c2i.baseURL, exchange, token)
    return ic.c2i.sendRequest(route, "GET")
}

// Utility methods (helpers)

func (ic *IntegrateData) isValidExchange(exchange string) bool {
    for _, ex := range ic.c2i.ExchangeTypes {
        if ex == exchange {
            return true
        }
    }
    return false
}

func (ic *IntegrateData) isValidTimeframe(timeframe string) bool {
    for _, tf := range ic.c2i.TimeframeTypes {
        if tf == timeframe {
            return true
        }
    }
    return false
}

func (ic *IntegrateData) getToken(exchange, tradingSymbol string) (string, error) {
    for _, symbol := range ic.c2i.Symbols {
        if symbol["segment"] == exchange && symbol["trading_symbol"] == tradingSymbol {
            return symbol["token"].(string), nil
        }
    }
    return "", fmt.Errorf("token not found for %s in symbols file", tradingSymbol)
}

func parseDate(dateStr string) time.Time {
    dt, err := time.Parse("020120061504", dateStr)
    if err != nil {
        return time.Time{}
    }
    return dt
}

func parseFields(data string) []string {
    return strings.Split(data, ",")
}

func toFloat(s string) float64 {
    f, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0.0
    }
    return f
}

func toInt(s string) int {
    i, err := strconv.Atoi(s)
    if err != nil {
        return 0
    }
    return i
}
