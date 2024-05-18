package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"
)

var ErrResponse = errors.New("internal server ")

type ServerResponse struct {
	Time time.Time `json:"time"`
}

func DefaultClient() *Client {
	c, err := NewClient("http://localhost:8085")
	if err != nil {
		zap.L().Panic("error init crypto_polymath client", zap.Error(err))
	}
	return c
}

type Client struct {
	httpClient *http.Client
	hostUrl    *url.URL
}

func NewClient(host string) (*Client, error) {
	hostUrl, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "parse host")
	}
	return &Client{
		hostUrl:    hostUrl,
		httpClient: http.DefaultClient,
	}, nil
}

func (c *Client) SetHttpClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

func (c *Client) Server(ctx context.Context) (ServerResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/server", c.hostUrl.String()),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ServerResponse{}, errors.Wrap(err, "send price request")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusInternalServerError {
		return ServerResponse{}, ErrResponse
	}
	var result ServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ServerResponse{}, err
	}

	return result, nil
}
