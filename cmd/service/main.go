// Package main contains the entrypoint of the binary that handles the service
// that tracks watch activity
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/Nivl/trakt-netflix/internal/activitytracker"
	"github.com/Nivl/trakt-netflix/internal/netflix"
	"github.com/Nivl/trakt-netflix/internal/slack"
	"github.com/Nivl/trakt-netflix/internal/trakt"
	"github.com/Nivl/trakt-netflix/internal/ui"
	"github.com/robfig/cron"
	"github.com/sethvargo/go-envconfig"
)

type appConfig struct {
	Trakt     trakt.ClientConfig `env:",prefix=TRAKT_"`
	Slack     slack.Config       `env:",prefix=SLACK_"`
	Netflix   netflix.Config     `env:",prefix=NETFLIX_"`
	CronSpecs string             `env:"CRON_SPECS,default=@hourly"`
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "something went wrong", "error", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
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

	if !traktClient.IsAuthenticated() {
		if err = ui.Authenticate(ctx, traktClient); err != nil {
			return fmt.Errorf("authenticate: %w", err)
		}
	}

	c := activitytracker.New(traktClient, netflixClient, slackClient)
	slog.InfoContext(ctx, "Trakt info: starting")

	crn := cron.New()
	err = crn.AddFunc(cfg.CronSpecs, func() {
		processCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		process(processCtx, c)
	})
	if err != nil {
		return fmt.Errorf("setup cron: %w", err)
	}
	crn.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.InfoContext(ctx, "Trakt info: stopping")

	crn.Stop()
	return nil
}

func process(ctx context.Context, c *activitytracker.Client) {
	if err := c.Run(ctx); err != nil {
		slog.InfoContext(ctx, "An error occurred during a run", "error", err)
		return
	}
}
