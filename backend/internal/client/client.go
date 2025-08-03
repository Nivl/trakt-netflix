package client

import (
	"net/http"
	"time"

	"github.com/Nivl/trakt-netflix/internal/trakt"
)

// Client represents a client to interact with external services
type Client struct {
	http          *http.Client
	slackWebhooks []string
	history       *History
	traktClient   *trakt.Client
}

// New returns a new Client
func New(slackWebhooks []string, history *History, traktClient *trakt.Client) *Client {
	return &Client{
		slackWebhooks: slackWebhooks,
		history:       history,
		traktClient:   traktClient,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
