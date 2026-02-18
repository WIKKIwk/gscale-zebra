package scaleapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type readingResponse struct {
	OK     bool     `json:"ok"`
	Weight *float64 `json:"weight"`
	Unit   string   `json:"unit"`
	Stable *bool    `json:"stable"`
	Error  string   `json:"error"`
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSpace(baseURL),
		http:    &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *Client) WaitStablePositive(ctx context.Context, timeout, pollInterval time.Duration) (float64, string, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if pollInterval <= 0 {
		pollInterval = 220 * time.Millisecond
	}

	deadline := time.Now().Add(timeout)
	var lastWeight float64
	var haveLast bool
	stableCount := 0

	for {
		if time.Now().After(deadline) {
			return 0, "", fmt.Errorf("scale qty timeout (%s)", timeout)
		}
		select {
		case <-ctx.Done():
			return 0, "", ctx.Err()
		default:
		}

		resp, err := c.readOnce(ctx)
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}
		if strings.TrimSpace(resp.Error) != "" {
			time.Sleep(pollInterval)
			continue
		}
		if resp.Weight == nil || *resp.Weight <= 0 {
			haveLast = false
			stableCount = 0
			time.Sleep(pollInterval)
			continue
		}

		w := *resp.Weight
		if resp.Stable != nil && *resp.Stable {
			return w, normalizeUnit(resp.Unit), nil
		}

		if haveLast && almostEqual(lastWeight, w, 0.001) {
			stableCount++
		} else {
			stableCount = 1
		}
		haveLast = true
		lastWeight = w

		if stableCount >= 4 {
			return w, normalizeUnit(resp.Unit), nil
		}
		time.Sleep(pollInterval)
	}
}

func (c *Client) readOnce(ctx context.Context) (readingResponse, error) {
	if c == nil || strings.TrimSpace(c.baseURL) == "" {
		return readingResponse{}, fmt.Errorf("scale api url bo'sh")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return readingResponse{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return readingResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return readingResponse{}, fmt.Errorf("scale http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload readingResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return readingResponse{}, fmt.Errorf("scale json xato: %w", err)
	}
	return payload, nil
}

func normalizeUnit(v string) string {
	u := strings.TrimSpace(v)
	if u == "" {
		return "kg"
	}
	return u
}

func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}
