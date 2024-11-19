package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Nivl/trakt-netflix/internal/client"
	"github.com/robfig/cron"
	"github.com/sethvargo/go-envconfig"
)

type appConfig struct {
	Netflix       client.NetflixConfig `env:",prefix=NETFLIX_"`
	Trakt         client.TraktConfig   `env:",prefix=TRAKT_"`
	SlackWebhooks []string             `env:"SLACK_WEBHOOKS"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("something went wrong", "error", err.Error())
		os.Exit(1)
	}
}

func run() (err error) {
	ctx := context.Background()
	var cfg appConfig
	if err = envconfig.Process(ctx, &cfg); err != nil {
		return fmt.Errorf("couldn't parse the env: %w", err)
	}

	c := client.New(cfg.SlackWebhooks)
	slog.Info("Trakt info: starting")

	currentHistory := client.NewHistory()
	if err = currentHistory.Load(); err != nil {
		slog.Warn("could not load history", "error", err.Error())
	}

	crn := cron.New()
	err = crn.AddFunc("@hourly", func() { process(&cfg, c, currentHistory) })
	if err != nil {
		return fmt.Errorf("could not setup cron: %w", err)
	}
	crn.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
	slog.Info("Trakt info: stopping")

	if err = currentHistory.Write(); err != nil {
		slog.Warn("could not write history", "error", err.Error())
	}
	crn.Stop()
	return nil
}

func process(cfg *appConfig, c *client.Client, currentHistory *client.History) {
	h, err := c.FetchNetflixHistory(cfg.Netflix, currentHistory)
	if err != nil {
		slog.Info("could not fetch shows from Netflix", "error", err)
		return
	}
	c.MarkAsWatched(cfg.Trakt, h)
}
