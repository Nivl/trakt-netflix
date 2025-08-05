package netflix

const HistorySize = 20

// Config contains the configuration needed for Netflix
type Config struct {
	AccountID string `env:"ACCOUNT_ID"`
	Cookie    string `env:"COOKIE,required"`
	URL       string `env:"URL,default=https://www.netflix.com/viewingactivity"`
}

