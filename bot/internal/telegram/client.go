package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

type getFileResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		FilePath string `json:"file_path"`
	} `json:"result"`
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	InlineQuery   *InlineQuery   `json:"inline_query"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int64       `json:"message_id"`
	Text      string      `json:"text"`
	Photo     []PhotoSize `json:"photo"`
	Chat      Chat        `json:"chat"`
	From      User        `json:"from"`
}

type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int64  `json:"file_size"`
}

type InlineQuery struct {
	ID     string `json:"id"`
	Query  string `json:"query"`
	Offset string `json:"offset"`
	From   User   `json:"from"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
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
	SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat,omitempty"`
	CallbackData                 string `json:"callback_data,omitempty"`
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
	q.Set("allowed_updates", `["message","inline_query","callback_query"]`)

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
	_, err := c.SendMessageWithInlineKeyboardAndReturnID(ctx, chatID, text, nil)
	return err
}

func (c *Client) SendMessageWithInlineKeyboard(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) error {
	_, err := c.SendMessageWithInlineKeyboardAndReturnID(ctx, chatID, text, keyboard)
	return err
}

func (c *Client) SendMessageWithInlineKeyboardAndReturnID(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) (int64, error) {
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)
	if keyboard != nil {
		b, err := json.Marshal(keyboard)
		if err != nil {
			return 0, err
		}
		form.Set("reply_markup", string(b))
	}
	return c.callSendMessage(ctx, form)
}

func (c *Client) EditMessageText(ctx context.Context, chatID, messageID int64, text string, keyboard *InlineKeyboardMarkup) error {
	if messageID <= 0 {
		return fmt.Errorf("message_id noto'g'ri")
	}
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("message_id", strconv.FormatInt(messageID, 10))
	form.Set("text", text)
	if keyboard != nil {
		b, err := json.Marshal(keyboard)
		if err != nil {
			return err
		}
		form.Set("reply_markup", string(b))
	}
	return c.callAPI(ctx, "editMessageText", form)
}

func (c *Client) DeleteMessage(ctx context.Context, chatID, messageID int64) error {
	if messageID <= 0 {
		return nil
	}

	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("message_id", strconv.FormatInt(messageID, 10))
	return c.callAPI(ctx, "deleteMessage", form)
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

func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string) error {
	form := url.Values{}
	form.Set("callback_query_id", strings.TrimSpace(callbackQueryID))
	if strings.TrimSpace(text) != "" {
		form.Set("text", strings.TrimSpace(text))
	}
	return c.callAPI(ctx, "answerCallbackQuery", form)
}

func (c *Client) GetFilePath(ctx context.Context, fileID string) (string, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return "", fmt.Errorf("file_id bo'sh")
	}

	q := url.Values{}
	q.Set("file_id", fileID)
	u := fmt.Sprintf("%s/bot%s/getFile?%s", c.baseURL, c.token, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload getFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if !payload.OK {
		if strings.TrimSpace(payload.Description) == "" {
			payload.Description = "getFile OK=false"
		}
		return "", fmt.Errorf("telegram: %s", payload.Description)
	}

	filePath := strings.TrimSpace(payload.Result.FilePath)
	if filePath == "" {
		return "", fmt.Errorf("telegram: file_path bo'sh")
	}
	return filePath, nil
}

func (c *Client) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	filePath = strings.TrimLeft(strings.TrimSpace(filePath), "/")
	if filePath == "" {
		return nil, fmt.Errorf("file_path bo'sh")
	}

	u := fmt.Sprintf("%s/file/bot%s/%s", c.baseURL, c.token, filePath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram: file download status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("telegram: file bo'sh")
	}
	return data, nil
}

func (c *Client) SendDocument(ctx context.Context, chatID int64, filename string, content []byte, caption string) error {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return fmt.Errorf("filename bo'sh")
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	if err := w.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		return err
	}
	if strings.TrimSpace(caption) != "" {
		if err := w.WriteField("caption", strings.TrimSpace(caption)); err != nil {
			return err
		}
	}

	part, err := w.CreateFormFile("document", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(content); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	u := fmt.Sprintf("%s/bot%s/sendDocument", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

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
			payload.Description = "sendDocument OK=false"
		}
		return fmt.Errorf("telegram: %s", payload.Description)
	}
	return nil
}

func (c *Client) callSendMessage(ctx context.Context, form url.Values) (int64, error) {
	u := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var payload sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}
	if !payload.OK {
		if strings.TrimSpace(payload.Description) == "" {
			payload.Description = "sendMessage OK=false"
		}
		return 0, fmt.Errorf("telegram: %s", payload.Description)
	}
	return payload.Result.MessageID, nil
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
