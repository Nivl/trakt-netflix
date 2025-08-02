package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Nivl/trakt-netflix/internal/trakt"
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

	authCode, err := traktClient.GenerateAuthCode(ctx)
	if err != nil {
		return fmt.Errorf("generate auth code: %w", err)
	}

	fmt.Printf("Please open the following URL in your browser:\n%s\n", path.Join(authCode.VerificationURL, authCode.UserCode))
	fmt.Printf("You have %d seconds to complete the authentication...\n", authCode.ExpiresIn)

	tickerSecond := time.NewTicker(1 * time.Second)
	tickerRetry := time.NewTicker(authCode.IntervalInSecs * time.Second)

	count := authCode.ExpiresIn
	fmt.Print(count)

	for {
		select {
		case <-tickerSecond.C:
			count--
			if count <= 0 {
				cancel()
				return fmt.Errorf("authentication timed out. Please try again")
			}
			fmt.Printf("\r\033[2K%d", count)
		case <-tickerRetry.C:
			_, err := traktClient.GetAccessToken(ctx, authCode.DeviceCode)
			if errors.Is(err, trakt.ErrPendingAuthorization) {
				continue
			}

			if err != nil {
				return fmt.Errorf("get access token: %w", err)
			}

			fmt.Println("\r\033[2KAuthentication successful")
			return nil
		}
	}
}
