package erp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

type Item struct {
	Name     string
	ItemCode string
	ItemName string
}

type listItemsResponse struct {
	Data []struct {
		Name     string `json:"name"`
		ItemCode string `json:"item_code"`
		ItemName string `json:"item_name"`
	} `json:"data"`
}

func New(baseURL, apiKey, apiSecret string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:    strings.TrimSpace(apiKey),
		apiSecret: strings.TrimSpace(apiSecret),
		http:      &http.Client{Timeout: 12 * time.Second},
	}
}

func (c *Client) CheckConnection(ctx context.Context) (string, error) {
	endpoint := c.baseURL + "/api/method/frappe.auth.get_logged_user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	c.setAuthHeader(req)

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

func (c *Client) SearchItems(ctx context.Context, query string, limit int) ([]Item, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	q := url.Values{}
	q.Set("fields", `[`+"\"name\",\"item_code\",\"item_name\""+`]`)
	q.Set("limit_page_length", strconv.Itoa(limit))
	q.Set("order_by", "modified desc")

	query = strings.TrimSpace(query)
	if query != "" {
		pattern := "%" + query + "%"
		orFilters := [][]interface{}{
			{"Item", "item_code", "like", pattern},
			{"Item", "item_name", "like", pattern},
			{"Item", "name", "like", pattern},
		}
		b, _ := json.Marshal(orFilters)
		q.Set("or_filters", string(b))
	}

	endpoint := c.baseURL + "/api/resource/Item?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("erp item http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload listItemsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("erp item json parse xato: %w", err)
	}

	items := make([]Item, 0, len(payload.Data))
	for _, r := range payload.Data {
		code := strings.TrimSpace(r.ItemCode)
		if code == "" {
			code = strings.TrimSpace(r.Name)
		}
		name := strings.TrimSpace(r.ItemName)
		if name == "" {
			name = code
		}
		if code == "" {
			continue
		}
		items = append(items, Item{
			Name:     strings.TrimSpace(r.Name),
			ItemCode: code,
			ItemName: name,
		})
	}

	return items, nil
}

func (c *Client) setAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", c.apiKey, c.apiSecret))
}
