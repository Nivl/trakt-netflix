package client

import (
	"github.com/Nivl/trakt-netflix/internal/netflix"
	"github.com/Nivl/trakt-netflix/internal/slack"
	"github.com/Nivl/trakt-netflix/internal/trakt"
)

// Client represents a client to interact with external services
type Client struct {
	slackWebhooks []string
	history       *History
	traktClient   *trakt.Client
	netflixClient *netflix.Client
	slackClient   *slack.Client
}

// New returns a new Client
func New(history *History, traktClient *trakt.Client, netflixClient *netflix.Client, slackClient *slack.Client) *Client {
	return &Client{
		slackClient:   slackClient,
		history:       history,
		traktClient:   traktClient,
		netflixClient: netflixClient,
	}
}
