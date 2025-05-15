package http

import (
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"net/http"
)

var _ httpSender = (*http.Client)(nil)

type httpSender interface {
	Do(r *http.Request) (*http.Response, error)
}

type Client struct {
	sender   httpSender
	maxRetry uint64
	logger   *zap.Logger
}

func NewClient(sender httpSender, maxRetry uint64, logger *zap.Logger) *Client {
	return &Client{sender: sender, maxRetry: maxRetry, logger: logger}
}

func (c *Client) Do(r *http.Request) (*http.Response, error) {
	var resp *http.Response
	err := backoff.Retry(func() error {
		var err error
		resp, err = c.sender.Do(r)
		if err != nil {
			c.logger.Debug("http request failed", zap.Error(err))
			return err
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), c.maxRetry))
	return resp, err
}
