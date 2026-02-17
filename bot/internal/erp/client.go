package erp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL   string
	apiKey    string
	apiSecret string
	http      *http.Client
}

type getUserResponse struct {
	Message string `json:"message"`
}

func New(baseURL, apiKey, apiSecret string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:    strings.TrimSpace(apiKey),
		apiSecret: strings.TrimSpace(apiSecret),
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) CheckConnection(ctx context.Context) (string, error) {
	endpoint := c.baseURL + "/api/method/frappe.auth.get_logged_user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", c.apiKey, c.apiSecret))

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("erp http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload getUserResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("erp json parse xato: %w", err)
	}
	if strings.TrimSpace(payload.Message) == "" {
		return "", fmt.Errorf("erp javob bo'sh")
	}
	return payload.Message, nil
}
