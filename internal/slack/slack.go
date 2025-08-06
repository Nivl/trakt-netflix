// Package slack provides a client for sending messages to Slack.
package slack

import (
	"context"
	"log/slog"

	"github.com/ashwanthkumar/slack-go-webhook"
)

// Config contains the configuration needed for Slack
type Config struct {
	WebhookURLs []string `env:"WEBHOOKS"`
}

// Client is a Slack client for sending messages.
type Client struct {
	webhookURLs []string
	Username    string
	IconEmoji   string
}

// NewClient creates a new Slack client.
func NewClient(cfg Config) *Client {
	return &Client{
		webhookURLs: cfg.WebhookURLs,
		Username:    "Trakt",
		IconEmoji:   ":strawberry:",
	}
}

// SendMessage sends a message to the registered Slack channels.
// Noop if the client is nil.
func (c *Client) SendMessage(ctx context.Context, msg string) {
	slog.InfoContext(ctx, msg)

	if c == nil {
		return
	}

	var firstError error
	for _, wh := range c.webhookURLs {
		payload := slack.Payload{
			Text:      msg,
			Username:  c.Username,
			IconEmoji: c.IconEmoji,
		}
		errs := slack.Send(wh, "", payload)
		if len(errs) > 0 {
			if firstError == nil {
				firstError = errs[0]
			}
			slog.ErrorContext(ctx, "failed sending slack message", "errors", errs, "webhookURL", wh)
		}
	}
}
