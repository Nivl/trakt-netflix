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
	watchActivityURL string
	cfg              Config
	http             Doer
}

func NewClient(cfg Config) (*Client, error) {
	return NewClientWithDoer(cfg, &http.Client{
		Timeout: 10 * time.Second,
	})
}

func NewClientWithDoer(cfg Config, doer Doer) (*Client, error) {
	u, err := url.JoinPath(cfg.URL, cfg.AccountID)
	if err != nil {
		return nil, fmt.Errorf("build watchActivityURL: %w", err)
	}

	return &Client{
		watchActivityURL: u,
		cfg:              cfg,
		http:             doer,
	}, nil
}

func (c *Client) request(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "NetflixId",
		Value: c.cfg.Cookie,
	})

	return c.http.Do(req)
}
