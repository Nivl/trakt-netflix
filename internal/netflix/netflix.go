// Package netflix provides a client for interacting with Netflix.
package netflix

// HistorySize is the maximum number of items to keep in the history.
const HistorySize = 20

// Config contains the configuration needed for Netflix
type Config struct {
	AccountID string `env:"ACCOUNT_ID"`
	Cookie    string `env:"COOKIE,required"`
	URL       string `env:"URL,default=https://www.netflix.com/viewingactivity"`
}
