package client

import (
	"net/http"
	"time"
)

// Client represents a client to interact with external services
type Client struct {
	http          *http.Client
	slackWebhooks []string
}

// New returns a new Client
func New(slackWebhooks []string) *Client {
	return &Client{
		slackWebhooks: slackWebhooks,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
