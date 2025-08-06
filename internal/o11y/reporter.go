// Package o11y provides observability utilities.
package o11y

import "context"

// Reporter is an interface for sending messages to an observability
// backend.
type Reporter interface {
	SendMessage(ctx context.Context, msg string)
}
