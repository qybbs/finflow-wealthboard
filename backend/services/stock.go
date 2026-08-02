package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type YahooFinanceResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

type StockService struct {
	cacheMutex 	sync.RWMutex
	priceCache 	map[string]float64
	lastUpdated map[string]time.Time
}

func NewStockService() *StockService {
	return &StockService{
		priceCache: make(map[string]float64),
		lastUpdated: make(map[string]time.Time),
	}
}

func (s *StockService) GetStockPrice(ticker string) (float64, error) {
	s.cacheMutex.RLock()
	lastTime, exists := s.lastUpdated[ticker]
	cachedPrice := s.priceCache[ticker]
	s.cacheMutex.RUnlock()

	if exists && time.Since(lastTime) < 5*time.Minute {
		fmt.Printf("[CACHE] Menggunakan harga cache untuk %s\n", ticker)
		return cachedPrice, nil
	}

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s.JK", ticker)

	fmt.Printf("[API] Menarik data langsung dari Yahoo Finance untuk %s...\n", ticker)

		// KODE PENGGANTI (Penyamaran User-Agent)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	// Kita "menyamar" menjadi browser Chrome di Windows
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	
	// Gunakan http.Client khusus untuk mengeksekusi request yang sudah disamarkan
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var yahooResp YahooFinanceResponse
	err = json.Unmarshal(bodyData, &yahooResp)
	if err != nil {
		return 0, err
	}

	if len(yahooResp.Chart.Result) == 0 {
		return 0, fmt.Errorf("data saham %s tidak ditemukan di Yahoo Finance", ticker)
	}

	currentPrice := yahooResp.Chart.Result[0].Meta.RegularMarketPrice

	s.cacheMutex.Lock()
	s.priceCache[ticker] = currentPrice
	s.lastUpdated[ticker] = time.Now()
	s.cacheMutex.Unlock()

	return currentPrice, nil
}