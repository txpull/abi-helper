// Package scanners provides the functionality to scan and interact with
// the 4byte.directory API, a signature database of 4byte function signatures.
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

// FourByteResult represents a single function signature entry.
type FourByteResult struct {
	ID        int       `json:"id"`
	Text      string    `json:"text_signature"`
	Hex       string    `json:"hex_signature"`
	CreatedAt time.Time `json:"created_at"`
}

// PageResponse is the structure of a response from the 4byte.directory API.
type PageResponse struct {
	Count    int              `json:"count"`
	Next     string           `json:"next"`
	Previous string           `json:"previous"`
	Results  []FourByteResult `json:"results"`
}

// FourByteProvider is a client for interacting with the 4byte.directory API.
type FourByteProvider struct {
	client     http.Client
	baseURL    string
	maxRetries int
	ctx        context.Context
}

// Option is a function that applies a configuration option to a FourByteProvider.
type Option func(*FourByteProvider)

// WithURL is an Option to set the base URL of the 4byte.directory API.
//
// Example usage:
//
//	provider := NewFourByteProvider(WithURL("https://www.4byte.directory/api/v1/signatures/"))
func WithURL(url string) Option {
	return func(p *FourByteProvider) {
		p.baseURL = url
	}
}

// WithMaxRetries is an Option to set the maximum number of retries on network failure.
//
// Example usage:
//
//	provider := NewFourByteProvider(WithMaxRetries(5))
func WithMaxRetries(maxRetries int) Option {
	return func(p *FourByteProvider) {
		p.maxRetries = maxRetries
	}
}

// WithContext is an Option to set the context of the FourByteProvider.
//
// Example usage:
//
//	ctx := context.Background()
//	provider := NewFourByteProvider(WithContext(ctx))
func WithContext(ctx context.Context) Option {
	return func(p *FourByteProvider) {
		p.ctx = ctx
	}
}

// NewFourByteProvider creates a new FourByteProvider instance with the provided Options.
// It defaults to using a background context if no context is provided,
// the default base URL is "https://www.4byte.directory/api/v1/signatures/",
// and the default maximum number of retries on network failure is 3.
//
// Example usage:
//
//	ctx := context.Background()
//	provider := NewFourByteProvider(WithContext(ctx), WithURL("https://www.4byte.directory/api/v1/signatures/"), WithMaxRetries(5))
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

// GetPage retrieves a page of function signature entries from the 4byte.directory API.
// If pageNum is 0, it retrieves the first page.
//
// Example usage:
//
//	pageResponse, err := provider.GetPage(1)
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
