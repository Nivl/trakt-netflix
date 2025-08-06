package netflix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nivl/trakt-netflix/internal/o11y"
	"github.com/Nivl/trakt-netflix/internal/pathutil"
)

// History represents the viewing history of a user.
type History struct {
	ItemsSearch map[string]struct{} `json:"search"`
	Items       []string            `json:"items"`
	NewActivity []*WatchActivity    `json:"-"`
}

// NewHistory creates a new History instance, and loads the initial
// data stored on disk.
func NewHistory() (*History, error) {
	h := &History{
		ItemsSearch: make(map[string]struct{}),
		Items:       []string{},
		NewActivity: []*WatchActivity{},
	}
	err := h.Load()
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	return h, nil
}

// Has checks if the history contains a specific item.
func (h *History) Has(item string) bool {
	_, ok := h.ItemsSearch[item]
	return ok
}

// Push adds a new item to the history.
func (h *History) Push(ctx context.Context, item string, r o11y.Reporter) {
	if h.Has(item) {
		return
	}

	if len(h.Items) >= HistorySize {
		delete(h.ItemsSearch, h.Items[0])
		h.Items = h.Items[1:]
	}

	h.Items = append(h.Items, item)
	h.ItemsSearch[item] = struct{}{}
	h.NewActivity = append(h.NewActivity, ParseTitle(ctx, item, r))
}

// Write saves the history to disk.
func (h *History) Write() error {
	dataFilePath := filepath.Join(pathutil.ConfigDir(), "history")

	data, err := json.Marshal(h)
	if err != nil {
		return fmt.Errorf("marshal the data: %w", err)
	}
	return os.WriteFile(dataFilePath, data, 0o600)
}

// Load loads the history from disk.
func (h *History) Load() error {
	dataFilePath := filepath.Join(pathutil.ConfigDir(), "history")

	// TODO(melvin): Use something more secure than ReadFile, to avoid
	// loading a huge file in memory.
	data, err := os.ReadFile(dataFilePath) //nolint:gosec // G304: file inclusion via variable is what we want here
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read the file: %w", err)
	}
	err = json.Unmarshal(data, h)
	if err != nil {
		return fmt.Errorf("unmarshal the data: %w", err)
	}
	return nil
}

// ClearNewActivity clears the new activity from the history.
func (h *History) ClearNewActivity() {
	h.NewActivity = []*WatchActivity{}
}
