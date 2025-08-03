package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Nivl/trakt-netflix/internal/client"
	"github.com/Nivl/trakt-netflix/internal/trakt"
	"github.com/robfig/cron"
	"github.com/sethvargo/go-envconfig"
)

type appConfig struct {
	Netflix       client.NetflixConfig `env:",prefix=NETFLIX_"`
	Trakt         trakt.ClientConfig   `env:",prefix=TRAKT_"`
	SlackWebhooks []string             `env:"SLACK_WEBHOOKS"`
	CronSpecs     string               `env:"CRON_SPECS,default=@hourly"`
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

	history := client.NewHistory()
	if err = history.Load(); err != nil {
		slog.Warn("could not load history", "error", err.Error())
	}

	traktClient, err := trakt.NewClient(cfg.Trakt)
	if err != nil {
		return fmt.Errorf("create trakt client: %w", err)
	}

	c := client.New(cfg.SlackWebhooks, history, traktClient)
	slog.Info("Trakt info: starting")

	crn := cron.New()
	err = crn.AddFunc(cfg.CronSpecs, func() { process(&cfg, c, history) })
	if err != nil {
		return fmt.Errorf("could not setup cron: %w", err)
	}
	crn.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("Trakt info: stopping")

	if err = history.Write(); err != nil {
		slog.Warn("could not write history", "error", err.Error())
	}
	crn.Stop()
	return nil
}

func process(cfg *appConfig, c *client.Client, history *client.History) {
	ctx := context.Background()

	err := c.FetchNetflixHistory(cfg.Netflix)
	if err != nil {
		slog.Info("could not fetch shows from Netflix", "error", err)
		return
	}
	c.MarkAsWatched(ctx)
	if err = history.Write(); err != nil {
		slog.Warn("could not write history", "error", err.Error())
	}
}
