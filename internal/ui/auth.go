package ui

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Nivl/trakt-netflix/internal/trakt"
)

// Authenticate prompts the user to authenticate with Trakt using the
// device code flow.
func Authenticate(ctx context.Context, traktClient *trakt.Client) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	authCode, err := traktClient.GenerateAuthCode(ctx)
	if err != nil {
		return fmt.Errorf("generate auth code: %w", err)
	}

	finalURL, err := url.JoinPath(authCode.VerificationURL, authCode.UserCode)
	if err != nil {
		finalURL = fmt.Sprintf("%s and enter the code: %s", authCode.VerificationURL, authCode.UserCode)
	}
	fmt.Printf("Please open the following URL in your browser:\n%s\n", finalURL)
	fmt.Printf("You have %d seconds to complete the authentication...\n", authCode.ExpiresInSecs)

	tickerSecond := time.NewTicker(1 * time.Second)
	defer tickerSecond.Stop()
	tickerRetry := time.NewTicker(time.Duration(authCode.IntervalInSecs) * time.Second)
	defer tickerRetry.Stop()

	count := authCode.ExpiresInSecs
	fmt.Print(count)

	for {
		select {
		case <-tickerSecond.C:
			count--
			if count <= 0 {
				cancel()
				return errors.New("authentication timed out. Please try again")
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
