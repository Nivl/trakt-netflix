package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Nivl/trakt-netflix/internal/activitytracker"
	"github.com/Nivl/trakt-netflix/internal/netflix"
	"github.com/Nivl/trakt-netflix/internal/slack"
	"github.com/Nivl/trakt-netflix/internal/trakt"
	"github.com/robfig/cron"
	"github.com/sethvargo/go-envconfig"
)

type appConfig struct {
	Netflix   netflix.Config     `env:",prefix=NETFLIX_"`
	Trakt     trakt.ClientConfig `env:",prefix=TRAKT_"`
	Slack     slack.Config       `env:",prefix=SLACK_"`
	CronSpecs string             `env:"CRON_SPECS,default=@hourly"`
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
		return fmt.Errorf("parse the env: %w", err)
	}

	traktClient, err := trakt.NewClient(cfg.Trakt)
	if err != nil {
		return fmt.Errorf("create trakt client: %w", err)
	}

	netflixClient, err := netflix.NewClient(cfg.Netflix)
	if err != nil {
		return fmt.Errorf("create netflix client: %w", err)
	}

	slackClient := slack.NewClient(cfg.Slack)

	c := activitytracker.New(traktClient, netflixClient, slackClient)
	slog.Info("Trakt info: starting")

	crn := cron.New()
	err = crn.AddFunc(cfg.CronSpecs, func() { process(c) })
	if err != nil {
		return fmt.Errorf("setup cron: %w", err)
	}
	crn.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("Trakt info: stopping")

	crn.Stop()
	return nil
}

func process(c *activitytracker.Client) {
	ctx := context.Background()
	if err := c.Run(ctx); err != nil {
		slog.Info("An error occurred during a run", "error", err)
		return
	}
}
