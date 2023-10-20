package client

import (
	"os"
)

// ConfigDir returns the path to the config directory
func ConfigDir() string {
	_, err := os.Stat("/.dockerenv")
	if err == nil {
		return "/config"
	}

	// fallback on current directory
	return ""
}
