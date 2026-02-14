package market

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nofx/hook"
	"strconv"
	"time"
)

const (
	baseURL = "https://fapi.binance.com"
)

type APIClient struct {
	client *http.Client
}

func NewAPIClient() *APIClient {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	hookRes := hook.HookExec[hook.SetHttpClientResult](hook.SET_HTTP_CLIENT, client)
	if hookRes != nil && hookRes.Error() == nil {
		log.Printf("Using HTTP client set by Hook")
		client = hookRes.GetResult()
	}

	return &APIClient{
		client: client,
	}
}

func (c *APIClient) GetExchangeInfo() (*ExchangeInfo, error) {
	url := fmt.Sprintf("%s/fapi/v1/exchangeInfo", baseURL)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var exchangeInfo ExchangeInfo
	err = json.Unmarshal(body, &exchangeInfo)
	if err != nil {
		return nil, err
	}

	return &exchangeInfo, nil
}

func (c *APIClient) GetKlines(symbol, interval string, limit int) ([]Kline, error) {
	url := fmt.Sprintf("%s/fapi/v1/klines", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("symbol", symbol)
	q.Add("interval", interval)
	q.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klineResponses []KlineResponse
	err = json.Unmarshal(body, &klineResponses)
	if err != nil {
		log.Printf("Failed to get K-line data, response content: %s", string(body))
		return nil, err
	}

	var klines []Kline
	for _, kr := range klineResponses {
		kline, err := parseKline(kr)
		if err != nil {
			log.Printf("Failed to parse K-line data: %v", err)
			continue
		}
		klines = append(klines, kline)
	}

	return klines, nil
}

func parseKline(kr KlineResponse) (Kline, error) {
	var kline Kline

	if len(kr) < 11 {
		return kline, fmt.Errorf("invalid kline data")
	}

	// Parse each field
	kline.OpenTime = int64(kr[0].(float64))

	// Safely parse string values with empty checks
	openStr, ok := kr[1].(string)
	if !ok || openStr == "" {
		return kline, fmt.Errorf("invalid open price value: %v", kr[1])
	}
	open, err := strconv.ParseFloat(openStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse open price: %w", err)
	}
	kline.Open = open

	highStr, ok := kr[2].(string)
	if !ok || highStr == "" {
		return kline, fmt.Errorf("invalid high price value: %v", kr[2])
	}
	high, err := strconv.ParseFloat(highStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse high price: %w", err)
	}
	kline.High = high

	lowStr, ok := kr[3].(string)
	if !ok || lowStr == "" {
		return kline, fmt.Errorf("invalid low price value: %v", kr[3])
	}
	low, err := strconv.ParseFloat(lowStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse low price: %w", err)
	}
	kline.Low = low

	closeStr, ok := kr[4].(string)
	if !ok || closeStr == "" {
		return kline, fmt.Errorf("invalid close price value: %v", kr[4])
	}
	close, err := strconv.ParseFloat(closeStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse close price: %w", err)
	}
	kline.Close = close

	volumeStr, ok := kr[5].(string)
	if !ok || volumeStr == "" {
		return kline, fmt.Errorf("invalid volume value: %v", kr[5])
	}
	volume, err := strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse volume: %w", err)
	}
	kline.Volume = volume

	kline.CloseTime = int64(kr[6].(float64))

	quoteVolumeStr, ok := kr[7].(string)
	if !ok || quoteVolumeStr == "" {
		return kline, fmt.Errorf("invalid quote volume value: %v", kr[7])
	}
	quoteVolume, err := strconv.ParseFloat(quoteVolumeStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse quote volume: %w", err)
	}
	kline.QuoteVolume = quoteVolume

	kline.Trades = int(kr[8].(float64))

	takerBuyBaseVolumeStr, ok := kr[9].(string)
	if !ok || takerBuyBaseVolumeStr == "" {
		return kline, fmt.Errorf("invalid taker buy base volume value: %v", kr[9])
	}
	takerBuyBaseVolume, err := strconv.ParseFloat(takerBuyBaseVolumeStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse taker buy base volume: %w", err)
	}
	kline.TakerBuyBaseVolume = takerBuyBaseVolume

	takerBuyQuoteVolumeStr, ok := kr[10].(string)
	if !ok || takerBuyQuoteVolumeStr == "" {
		return kline, fmt.Errorf("invalid taker buy quote volume value: %v", kr[10])
	}
	takerBuyQuoteVolume, err := strconv.ParseFloat(takerBuyQuoteVolumeStr, 64)
	if err != nil {
		return kline, fmt.Errorf("failed to parse taker buy quote volume: %w", err)
	}
	kline.TakerBuyQuoteVolume = takerBuyQuoteVolume

	return kline, nil
}

func (c *APIClient) GetCurrentPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("%s/fapi/v1/ticker/price", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Add("symbol", symbol)
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var ticker PriceTicker
	err = json.Unmarshal(body, &ticker)
	if err != nil {
		return 0, err
	}

	// Check if price is empty or invalid
	if ticker.Price == "" {
		return 0, fmt.Errorf("empty price received for symbol %s", symbol)
	}

	price, err := strconv.ParseFloat(ticker.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price '%s' for symbol %s: %w", ticker.Price, symbol, err)
	}

	return price, nil
}
