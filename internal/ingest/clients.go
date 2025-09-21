package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	Client *http.Client
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{Client: &http.Client{Timeout: timeout}}
}

func (c *HTTPClient) GetJSON(ctx context.Context, url string, out any) error {
	var lastErr error
	backoff := []time.Duration{300 * time.Millisecond, 600 * time.Millisecond, 1200 * time.Millisecond}

	for attempt := 0; attempt <= len(backoff); attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := c.Client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				body, _ := io.ReadAll(resp.Body)
				return json.Unmarshal(body, out)
			}
			// 4xx no se reintenta (normalmente)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return errors.New(resp.Status)
			}
			lastErr = errors.New(resp.Status)
		}
		// backoff si hay más intentos
		if attempt < len(backoff) {
			select {
			case <-time.After(backoff[attempt]):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return lastErr
}
