package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	apiKey      string
	redisClient *redis.Client
	httpClient  *http.Client
}

type FixerResponse struct {
	Success bool               `json:"success"`
	Query   FixerQuery         `json:"query"`
	Info    FixerInfo          `json:"info"`
	Result  float64            `json:"result"`
	Rates   map[string]float64 `json:"rates"`
}

type FixerQuery struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

type FixerInfo struct {
	Timestamp int64   `json:"timestamp"`
	Rate      float64 `json:"rate"`
}

func NewService(apiKey string, redisClient *redis.Client) *Service {
	return &Service{
		apiKey:      apiKey,
		redisClient: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ConvertToRUB конвертирует сумму из любой валюты в рубли
func (s *Service) ConvertToRUB(ctx context.Context, amount float64, fromCurrency string) (float64, error) {
	// Если уже рубли - возвращаем как есть
	if fromCurrency == "RUB" {
		return amount, nil
	}

	// Проверяем кеш (курсы кешируем на 1 час)
	cacheKey := fmt.Sprintf("exchange_rate:%s:RUB", fromCurrency)
	if cachedRate, err := s.redisClient.Get(ctx, cacheKey).Float64(); err == nil {
		return amount * cachedRate, nil
	}

	// Запрашиваем курс у Fixer.io
	rate, err := s.getExchangeRate(fromCurrency, "RUB")
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Кешируем курс на 1 час
	s.redisClient.Set(ctx, cacheKey, rate, time.Hour)

	return amount * rate, nil
}

// getExchangeRate получает курс конвертации через Fixer.io
func (s *Service) getExchangeRate(from, to string) (float64, error) {
	// Если API ключ не задан - используем fallback курсы
	if s.apiKey == "" {
		return s.getFallbackRate(from, to), nil
	}

	url := fmt.Sprintf("http://data.fixer.io/api/latest?access_key=%s&base=%s&symbols=%s",
		s.apiKey, from, to)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		// Fallback при ошибке сети
		return s.getFallbackRate(from, to), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback при ошибке API
		return s.getFallbackRate(from, to), nil
	}

	var fixerResp FixerResponse
	if err := json.NewDecoder(resp.Body).Decode(&fixerResp); err != nil {
		return s.getFallbackRate(from, to), nil
	}

	if !fixerResp.Success {
		return s.getFallbackRate(from, to), nil
	}

	rate, ok := fixerResp.Rates[to]
	if !ok {
		return s.getFallbackRate(from, to), nil
	}

	return rate, nil
}

// getFallbackRate возвращает примерные курсы (обновлено декабрь 2025)
func (s *Service) getFallbackRate(from, to string) float64 {
	// Базовые курсы к RUB (примерные)
	rates := map[string]float64{
		"USD": 95.0,  // 1 USD = 95 RUB
		"EUR": 105.0, // 1 EUR = 105 RUB
		"KZT": 0.20,  // 1 KZT = 0.20 RUB
		"AZN": 56.0,  // 1 AZN = 56 RUB
	}

	if to == "RUB" {
		if rate, ok := rates[from]; ok {
			return rate
		}
	}

	// Если комбинация неизвестна - возвращаем 1.0 (без конвертации)
	return 1.0
}

