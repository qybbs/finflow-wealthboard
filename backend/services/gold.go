package services

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GoldService akan menangani penarikan harga emas Antam
type GoldService struct {
	cacheMutex  sync.RWMutex
	priceCache  float64
	lastUpdated time.Time
}

func NewGoldService() *GoldService {
	return &GoldService{}
}

// GetGoldPrice mengambil harga emas per gram dari harga-emas.org
func (s *GoldService) GetGoldPrice() (float64, error) {
	s.cacheMutex.RLock()
	lastTime := s.lastUpdated
	cachedPrice := s.priceCache
	s.cacheMutex.RUnlock()

	// Cache 15 menit untuk harga emas (jarang berubah)
	if cachedPrice > 0 && time.Since(lastTime) < 15*time.Minute {
		return cachedPrice, nil
	}

	url := "https://harga-emas.org/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0 Safari/537.36")

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

	bodyString := string(bodyData)
	
	// Scrape simple menggunakan RegEx untuk harga Antam 1 Gram
	// Harga emas antam biasanya di table: "1 gram" ... "Rp. 1.200.000"
	// Akan disesuaikan jika regex gagal
	regex := regexp.MustCompile(`1 gram.*?Rp\.\s*([\d\.]+)`)
	matches := regex.FindStringSubmatch(bodyString)

	var currentPrice float64
	if len(matches) > 1 {
		priceStr := strings.ReplaceAll(matches[1], ".", "")
		currentPrice, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return 0, fmt.Errorf("gagal mem-parsing harga: %v", err)
		}
	} else {
		// Fallback sederhana jika struktur HTML berubah
		return 0, fmt.Errorf("gagal mengekstrak harga emas dari HTML")
	}

	s.cacheMutex.Lock()
	s.priceCache = currentPrice
	s.lastUpdated = time.Now()
	s.cacheMutex.Unlock()

	return currentPrice, nil
}
