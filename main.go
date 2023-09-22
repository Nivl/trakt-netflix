package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Nivl/trakt-netflix/internal/client"
	"github.com/sethvargo/go-envconfig"
)

type appConfig struct {
	Netflix       client.NetflixConfig `env:",prefix=NETFLIX_"`
	Trakt         client.TraktConfig   `env:",prefix=TRAKT_"`
	SlackWebhooks []string             `env:"SLACK_WEBHOOKS"`
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	ctx := context.Background()
	var cfg appConfig
	if err = envconfig.Process(ctx, &cfg); err != nil {
		return fmt.Errorf("couldn't parse the env: %w", err)
	}

	c := client.New(cfg.SlackWebhooks)
	defer func() {
		if err != nil {
			c.Report("Trakt error: " + err.Error())
		}
	}()

	h, err := c.FetchNetflixHistory(cfg.Netflix)
	if err != nil {
		return fmt.Errorf("could not fetch shows from Netlifx: %w", err)
	}
	c.MarkAsWatched(cfg.Trakt, h)
	return nil
}
