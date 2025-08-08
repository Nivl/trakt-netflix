// Package main contains the entry point of the binary that handles user authentication
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Nivl/trakt-netflix/internal/trakt"
	"github.com/Nivl/trakt-netflix/internal/ui"
	"github.com/sethvargo/go-envconfig"
)

type appConfig struct {
	Trakt trakt.ClientConfig `env:",prefix=TRAKT_"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var cfg appConfig
	if err = envconfig.Process(ctx, &cfg); err != nil {
		return fmt.Errorf("parse the env: %w", err)
	}

	traktClient, err := trakt.NewClient(cfg.Trakt)
	if err != nil {
		return fmt.Errorf("create trakt client: %w", err)
	}

	err = ui.Authenticate(ctx, traktClient)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	return nil
}
