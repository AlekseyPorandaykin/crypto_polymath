package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type Client struct {
	host       *url.URL
	httpClient *http.Client
}

func DefaultClient() *Client {
	c, err := NewClient("https://api.telegram.org")
	if err != nil {
		zap.L().Fatal("init telegram client", zap.Error(err))
		return nil
	}
	return c
}

func NewClient(hostUrl string) (*Client, error) {
	host, err := url.Parse(hostUrl)
	if err != nil {
		return nil, err
	}
	return &Client{host: host, httpClient: http.DefaultClient}, nil
}

func (c *Client) SendMessage(ctx context.Context, apiToken string, chatId string, text string) (bool, error) {
	urlReq := c.host.JoinPath(fmt.Sprintf("bot%s/sendMessage", apiToken))
	q := urlReq.Query()
	q.Set("chat_id", chatId)
	q.Set("text", text)
	urlReq.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlReq.String(),
		nil)
	if err != nil {
		return false, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	result := SendMessageResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Ok, nil
}
