package integrate

import (
	"adapter-project/structs"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type IntegrateDataBeta struct {
	c2i     *LocalConnect
	logging bool
}

func NewIntegrateDataBeta(c2i *LocalConnect, logging bool) *IntegrateDataBeta {
	return &IntegrateDataBeta{
		logging: logging,
		c2i:     c2i,
	}
}

func findToken(symbols []structs.Symbol, exchange, tradingSymbol string) (string, error) {
	exchange = strings.TrimSpace(strings.ToUpper(exchange))
	tradingSymbol = strings.TrimSpace(strings.ToUpper(tradingSymbol))
	for _, s := range symbols {
		if strings.ToUpper(strings.TrimSpace(s.Segment)) == exchange &&
			strings.ToUpper(strings.TrimSpace(s.TradingSymbol)) == tradingSymbol {
			return s.Token, nil
		}
	}
	return "", fmt.Errorf("token not found for %s/%s in symbols file", exchange, tradingSymbol)
}

// HistoricalDataBeta streams historical data as a channel, similar to Python generator
func HistoricalDataBeta(io *IntegrateDataBeta, exchange, tradingSymbol, timeframe string, start, end time.Time) (<-chan map[string]interface{}, error) {
	if !contains(io.c2i.ExchangeTypes, exchange) {
		return nil, errors.New("Invalid exchange type")
	}
	if !contains(io.c2i.TimeframeTypes, timeframe) {
		return nil, errors.New("Invalid timeframe")
	}
	token, err := findToken(io.c2i.Symbol, exchange, tradingSymbol)
	if err != nil {
		return nil, fmt.Errorf("Token not found for %s in symbols file", tradingSymbol)
	}
	tokenInt, err := strconv.Atoi(token)
	if err != nil {
		return nil, fmt.Errorf("Invalid token format for %s/%s: %v", exchange, tradingSymbol, err)
	}
	url := fmt.Sprintf("https://data.definedgesecurities.com/sds/history/%s/%d/%s/%s/%s",
		exchange,
		tokenInt,
		timeframe,
		start.Format("020120061504")[:12],
		end.Format("020120061504")[:12],
	)
	resp, err := io.c2i.sendRequest(url, "", "GET", nil, nil, nil, nil, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("Error fetching historical data: %w", err)
	}
	csvData, ok := resp["data"].(string)
	if !ok || csvData == "" {
		return nil, errors.New("Unexpected response format: no CSV data")
	}
	ch := make(chan map[string]interface{})
	go func() {
		defer close(ch)
		lines := strings.Split(csvData, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			data := strings.Split(line, ",")
			if len(data) == 7 {
				// day/minute: DateTime, Open, High, Low, Close, Volume, OI
				dt, err := time.Parse("020120061504", data[0])
				if err != nil {
					continue
				}
				ch <- map[string]interface{}{
					"datetime": dt,
					"open":     parseFloat(data[1]),
					"high":     parseFloat(data[2]),
					"low":      parseFloat(data[3]),
					"close":    parseFloat(data[4]),
					"volume":   parseInt64(data[5]),
					"oi":       parseInt64(data[6]),
				}
			} else if len(data) == 4 {
				// tick: UTC, LTP, LTQ, OI
				ch <- map[string]interface{}{
					"utc": data[0],
					"ltp": parseFloat(data[1]),
					"ltq": parseFloat(data[2]),
					"oi":  parseFloat(data[3]),
				}
			} else if len(data) == 6 {
				// day/minute without OI
				dt, err := time.Parse("020120061504", data[0])
				if err != nil {
					continue
				}
				ch <- map[string]interface{}{
					"datetime": dt,
					"open":     parseFloat(data[1]),
					"high":     parseFloat(data[2]),
					"low":      parseFloat(data[3]),
					"close":    parseFloat(data[4]),
					"volume":   parseInt64(data[5]),
				}
			}
		}
	}()
	return ch, nil
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func parseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func parseTime(s string) time.Time {
	t, err := time.Parse("020120061504", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func QuotesBeta(io *IntegrateDataBeta, exchange, tradingSymbol string) (map[string]interface{}, error) {
	if !contains(io.c2i.ExchangeTypes, exchange) {
		return nil, errors.New("invalid exchange type")
	}
	token, err := findToken(io.c2i.Symbol, exchange, tradingSymbol)
	if err != nil {
		return nil, err
	}
	route := fmt.Sprintf("quotes/%s/%s", exchange, token)
	return io.c2i.sendRequest(
		io.c2i.BaseURL,
		route,
		"GET",
		nil, nil, nil, nil, nil,
	)
}

func SecurityInformationBeta(io *IntegrateDataBeta, exchange string, tradingSymbol string) (map[string]interface{}, error) {
	if !contains(io.c2i.ExchangeTypes, exchange) {
		return nil, errors.New("invalid exchange type")
	}
	token, err := findToken(io.c2i.Symbol, exchange, tradingSymbol)
	if err != nil {
		return nil, err
	}
	route := fmt.Sprintf("securityinfo/%s/%s", exchange, token)
	return io.c2i.sendRequest(
		io.c2i.BaseURL,
		route,
		"GET",
		nil, nil, nil, nil, nil,
	)
}
