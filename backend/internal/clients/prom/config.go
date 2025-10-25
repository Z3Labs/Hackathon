package prom

import (
	"net/http"
	"time"
)

type VMClientConfig struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewDefaultConfig(baseURL string) *VMClientConfig {
	return &VMClientConfig{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *VMClientConfig) WithHTTPClient(client *http.Client) *VMClientConfig {
	c.HTTPClient = client
	return c
}

func (c *VMClientConfig) WithTimeout(timeout time.Duration) *VMClientConfig {
	c.HTTPClient.Timeout = timeout
	return c
}
