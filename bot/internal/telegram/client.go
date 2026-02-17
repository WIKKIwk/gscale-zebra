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
	OK          bool     `json:"ok"`
	Description string   `json:"description"`
	Result      []Update `json:"result"`
}

type apiOKResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

type Update struct {
	UpdateID    int64        `json:"update_id"`
	Message     *Message     `json:"message"`
	InlineQuery *InlineQuery `json:"inline_query"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	From User   `json:"from"`
}

type InlineQuery struct {
	ID     string `json:"id"`
	Query  string `json:"query"`
	Offset string `json:"offset"`
	From   User   `json:"from"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text                         string `json:"text"`
	SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat"`
}

type InlineQueryResultArticle struct {
	Type                string                  `json:"type"`
	ID                  string                  `json:"id"`
	Title               string                  `json:"title"`
	Description         string                  `json:"description,omitempty"`
	InputMessageContent InputTextMessageContent `json:"input_message_content"`
}

type InputTextMessageContent struct {
	MessageText string `json:"message_text"`
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
	q.Set("allowed_updates", `["message","inline_query"]`)

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
		if strings.TrimSpace(payload.Description) == "" {
			payload.Description = "getUpdates OK=false"
		}
		return nil, fmt.Errorf("telegram: %s", payload.Description)
	}
	return payload.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.SendMessageWithInlineKeyboard(ctx, chatID, text, nil)
}

func (c *Client) SendMessageWithInlineKeyboard(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) error {
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)
	if keyboard != nil {
		b, err := json.Marshal(keyboard)
		if err != nil {
			return err
		}
		form.Set("reply_markup", string(b))
	}
	return c.callAPI(ctx, "sendMessage", form)
}

func (c *Client) AnswerInlineQuery(ctx context.Context, inlineQueryID string, results []InlineQueryResultArticle, cacheSeconds int) error {
	if cacheSeconds < 0 {
		cacheSeconds = 0
	}
	b, err := json.Marshal(results)
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Set("inline_query_id", inlineQueryID)
	form.Set("results", string(b))
	form.Set("cache_time", strconv.Itoa(cacheSeconds))
	form.Set("is_personal", "true")
	return c.callAPI(ctx, "answerInlineQuery", form)
}

func (c *Client) callAPI(ctx context.Context, method string, form url.Values) error {
	u := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)
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

	var payload apiOKResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if !payload.OK {
		if strings.TrimSpace(payload.Description) == "" {
			payload.Description = method + " OK=false"
		}
		return fmt.Errorf("telegram: %s", payload.Description)
	}
	return nil
}
