package client

import (
	"log"

	"github.com/ashwanthkumar/slack-go-webhook"
)

// Report reports a message to Slack
func (c *Client) Report(msg string) {
	for _, wh := range c.slackWebhooks {
		payload := slack.Payload{
			Text:      msg,
			Username:  "Trakt",
			IconEmoji: ":strawberry:",
		}
		err := slack.Send(wh, "", payload)
		if len(err) > 0 {
			log.Printf("error: %s\n", err)
		}
	}
}
