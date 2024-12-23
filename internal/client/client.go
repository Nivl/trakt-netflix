package client

import (
	"net/http"
	"time"
)

// Client represents a client to interact with external services
type Client struct {
	http          *http.Client
	slackWebhooks []string
	history       *History
}

// New returns a new Client
func New(slackWebhooks []string, history *History) *Client {
	return &Client{
		slackWebhooks: slackWebhooks,
		history:       history,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
