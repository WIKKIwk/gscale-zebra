package qtyfile

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

type Client struct {
	path string
}

type snapshot struct {
	Weight    *float64 `json:"weight"`
	Unit      string   `json:"unit"`
	Stable    *bool    `json:"stable"`
	Error     string   `json:"error"`
	UpdatedAt string   `json:"updated_at"`
}

func New(path string) *Client {
	return &Client{path: strings.TrimSpace(path)}
}

func (c *Client) WaitStablePositive(ctx context.Context, timeout, pollInterval time.Duration) (float64, string, error) {
	if c == nil || strings.TrimSpace(c.path) == "" {
		return 0, "", fmt.Errorf("qty file path bo'sh")
	}
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

		s, err := c.readSnapshot()
		if err != nil {
			time.Sleep(pollInterval)
			continue
		}
		if strings.TrimSpace(s.Error) != "" {
			haveLast = false
			stableCount = 0
			time.Sleep(pollInterval)
			continue
		}
		if s.Weight == nil || *s.Weight <= 0 {
			haveLast = false
			stableCount = 0
			time.Sleep(pollInterval)
			continue
		}
		if !isFreshSnapshot(s.UpdatedAt, 4*time.Second) {
			time.Sleep(pollInterval)
			continue
		}

		w := *s.Weight
		if s.Stable != nil && *s.Stable {
			return w, normalizeUnit(s.Unit), nil
		}

		if haveLast && almostEqual(lastWeight, w, 0.001) {
			stableCount++
		} else {
			stableCount = 1
		}
		haveLast = true
		lastWeight = w

		if stableCount >= 4 {
			return w, normalizeUnit(s.Unit), nil
		}
		time.Sleep(pollInterval)
	}
}

func (c *Client) readSnapshot() (snapshot, error) {
	b, err := os.ReadFile(c.path)
	if err != nil {
		return snapshot{}, err
	}
	var s snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return snapshot{}, err
	}
	return s, nil
}

func isFreshSnapshot(updated string, maxAge time.Duration) bool {
	updated = strings.TrimSpace(updated)
	if updated == "" {
		return false
	}
	ts, err := time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return false
	}
	age := time.Since(ts)
	if age < 0 {
		age = 0
	}
	return age <= maxAge
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
