package scanners

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Result structure
type FourByteResult struct {
	ID        int       `json:"id"`
	Text      string    `json:"text_signature"`
	Hex       string    `json:"hex_signature"`
	CreatedAt time.Time `json:"created_at"`
}

// PageResponse structure
type PageResponse struct {
	Count    int              `json:"count"`
	Next     string           `json:"next"`
	Previous string           `json:"previous"`
	Results  []FourByteResult `json:"results"`
}

type FourByteProvider struct {
	client     http.Client
	baseURL    string
	maxRetries int
	ctx        context.Context
}

type Option func(*FourByteProvider)

func WithURL(url string) Option {
	return func(p *FourByteProvider) {
		p.baseURL = url
	}
}

func WithMaxRetries(maxRetries int) Option {
	return func(p *FourByteProvider) {
		p.maxRetries = maxRetries
	}
}

func WithContext(ctx context.Context) Option {
	return func(p *FourByteProvider) {
		p.ctx = ctx
	}
}

func NewFourByteProvider(opts ...Option) *FourByteProvider {
	provider := &FourByteProvider{
		client:     http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://www.4byte.directory/api/v1/signatures/",
		maxRetries: 3,                    // Default value
		ctx:        context.Background(), // Default value
	}

	for _, opt := range opts {
		opt(provider)
	}

	return provider
}

func (p *FourByteProvider) GetPage(pageNum uint64) (*PageResponse, error) {
	pageUrl := fmt.Sprintf("%s?page=%d", p.baseURL, pageNum)
	if pageNum == 0 {
		pageUrl = p.baseURL
	}

	req, err := http.NewRequestWithContext(p.ctx, http.MethodGet, pageUrl, nil)
	if err != nil {
		zap.L().Error("Failed to create new request", zap.Error(err))
		return nil, err
	}

	var lastErr error
	for i := 0; i < p.maxRetries; i++ {
		resp, err := p.client.Do(req)
		if err != nil {
			zap.L().Error("Failed to do request", zap.Error(err))
			lastErr = err
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = errors.New("unexpected http GET status: " + resp.Status)
			zap.L().Error("Unexpected HTTP status", zap.String("status", resp.Status))
			continue
		}

		pageResponse := new(PageResponse)
		if err := json.NewDecoder(resp.Body).Decode(pageResponse); err != nil {
			zap.L().Error("Failed to decode response body", zap.Error(err))
			lastErr = err
			continue
		}

		return pageResponse, nil
	}

	if lastErr != nil {
		zap.L().Error("Failed to get page after retries", zap.Uint64("page number", pageNum), zap.Error(lastErr))
	}
	return nil, lastErr
}
