package slack

import (
	"log/slog"

	"github.com/ashwanthkumar/slack-go-webhook"
)

// Config contains the configuration needed for Slack
type Config struct {
	WebhookURLs []string `env:"WEBHOOKS"`
}

type Client struct {
	webhookURLs []string
	Username    string
	IconEmoji   string
}

func NewClient(cfg Config) *Client {
	return &Client{
		webhookURLs: cfg.WebhookURLs,
		Username:    "Trakt",
		IconEmoji:   ":strawberry:",
	}
}

func (c *Client) SendMessage(msg string) error {
	slog.Info(msg)

	if c == nil {
		return nil
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
			slog.Error("failed sending slack message", "errors", errs, "webhookURL", wh)
		}
	}
	return firstError
}
