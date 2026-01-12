package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
}

func NewSecureClient(insecureSkipVerify bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Client{
		Client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}
