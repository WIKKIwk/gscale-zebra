package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

type updatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type sendMessageResponse struct {
	OK bool `json:"ok"`
}

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	From User   `json:"from"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

func New(token string) *Client {
	return &Client{
		token:   strings.TrimSpace(token),
		baseURL: "https://api.telegram.org",
		http:    &http.Client{Timeout: 70 * time.Second},
	}
}

func (c *Client) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	q := url.Values{}
	q.Set("offset", strconv.FormatInt(offset, 10))
	q.Set("timeout", strconv.Itoa(timeoutSec))
	q.Set("allowed_updates", `["message"]`)

	u := fmt.Sprintf("%s/bot%s/getUpdates?%s", c.baseURL, c.token, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload updatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("telegram getUpdates OK=false")
	}
	return payload.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)

	u := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var payload sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if !payload.OK {
		return fmt.Errorf("telegram sendMessage OK=false")
	}
	return nil
}
