package netflix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Doer is an interface that wraps the Do method of http.Client.
//
//go:generate mockgen -destination=../mocks/doer.go -package=mocks github.com/Nivl/trakt-netflix/internal/netflix Doer
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a struct that represents a client for interacting with Netflix.
type Client struct {
	HTTP             Doer
	History          *History
	WatchActivityURL string
	Cookie           string
}

// NewClient creates a new Client for interacting with Netflix.
func NewClient(cfg Config) (*Client, error) {
	u, err := url.JoinPath(cfg.URL, cfg.AccountID)
	if err != nil {
		return nil, fmt.Errorf("build watchActivityURL: %w", err)
	}

	watchHistory, err := NewHistory()
	if err != nil {
		return nil, fmt.Errorf("create history: %w", err)
	}
	return &Client{
		WatchActivityURL: u,
		Cookie:           cfg.Cookie,
		History:          watchHistory,
		HTTP: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) request(ctx context.Context, targetURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "NetflixId",
		Value: c.Cookie,
	})

	return c.HTTP.Do(req)
}
