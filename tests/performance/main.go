package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var simpleRequests = []string{
	"/api/server",
	"/api/prices/exchange/bybit",
	"/api/symbols/bybit/future",
	"/api/symbols/bybit/spot",
	"/api/price/bybit/BTCUSDT",
	"/api/prices/symbol/BTCUSDT",
	"/api/exchange/bybit/BTCUSDT",
}

var requestsWithSymbolParam = []string{
	"/api/candlestick/bybit/%s/H/1",

	"/api/indicator/bybit/%s/H/1/Trend/10",
	"/api/indicator/bybit/%s/H/1/MA/10",
	"/api/indicator/bybit/%s/H/1/EMA/10",
	"/api/indicator/bybit/%s/H/1/TypeCandle/1",
	"/api/indicator/bybit/%s/H/1/VolatilityCandlePercent/1",
	"/api/indicator/bybit/%s/H/1/PriceChanges/10",
	"/api/indicator/bybit/%s/H/1/StochasticMainLine/10",

	"/api/candle-indicator/bybit/%s/H/1/HeikenAshi",

	"/api/analysis/bybit/%s/H/1/TrendByMA/10/10",
	"/api/analysis/bybit/%s/H/1/TrendByEMA/10/10",
	"/api/analysis/bybit/%s/H/1/RatioCandleToMA/10/1",
	"/api/analysis/bybit/%s/H/1/RatioCandleToEMA/10/1",
	"/api/analysis/bybit/%s/H/1/RSI/10/10",
	"/api/analysis/bybit/%s/H/1/MACDMainLine/26/1",
	"/api/analysis/bybit/%s/H/1/MACDSignalLine/26/1",
	"/api/analysis/bybit/%s/H/1/MACDSHistogram/26/1",
	"/api/analysis/bybit/%s/H/1/StochasticSignalLine/10/3",
}

var host = "http://37.1.216.169"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	symbols := fetchAllSymbols()
	wg := sync.WaitGroup{}
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(50)
	for _, simpleRequest := range simpleRequests {
		wg.Add(1)
		go func(req string) {
			wg.Done()
			sendManyRequests(ctx, simpleRequest, 100)
		}(simpleRequest)
	}
	for _, requestWithSymbolParam := range requestsWithSymbolParam {
		for _, symbol := range symbols {
			wg.Add(1)
			s := symbol
			urlQ := requestWithSymbolParam
			g.Go(func() error {
				defer wg.Done()
				sendSymbolRequest(ctx, fmt.Sprintf(urlQ, s))
				return nil
			})
		}
	}

	wg.Wait()
}

func sendSymbolRequest(ctx context.Context, url string) {
	sendManyRequests(ctx, url, 100)
}

func sendManyRequests(ctx context.Context, url string, limit int) {
	for i := 0; i < limit; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			if err := sendRequest(ctx, url); err != nil {
				fmt.Printf("Error sending request: %v\n", err)
			}
		}
	}
}

func sendRequest(ctx context.Context, url string) error {
	start := time.Now()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", host, url), nil)
	if err != nil {
		return err
	}
	req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("Request=%s | Response status: %s, time=%s \n", url, resp.Status, time.Since(start).String())
	return nil
}

func fetchAllSymbols() []string {
	data := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "XRPUSDT", "LTCUSDT", "DOGEUSDT", "SOLUSDT", "DOTUSDT", "MATICUSDT", "TRXUSDT"}
	resp, err := http.DefaultClient.Get("/api/symbols/bybit/future")
	if err != nil {
		return data
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return data
	}
	type dataDto struct {
		Symbol string `json:"symbol"`
	}
	respDataDto := make([]dataDto, 0, 1000)
	if err := json.NewDecoder(resp.Body).Decode(&respDataDto); err != nil {
		return data
	}
	if len(respDataDto) == 0 {
		return data
	}
	data = make([]string, 0, len(respDataDto))
	for _, item := range respDataDto {
		data = append(data, item.Symbol)
	}
	return data
}
