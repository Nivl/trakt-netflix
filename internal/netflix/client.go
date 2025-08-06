package netflix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

//go:generate mockgen -destination=../mocks/doer.go -package=mocks github.com/Nivl/trakt-netflix/internal/netflix Doer
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	HTTP             Doer
	History          *History
	WatchActivityURL string
	Cookie           string
}

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

func (c *Client) request(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "NetflixId",
		Value: c.Cookie,
	})

	return c.HTTP.Do(req)
}
