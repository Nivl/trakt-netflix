package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nivl/trakt-netflix/internal/pathutil"
)

type History struct {
	ItemsSearch map[string]struct{} `json:"search"`
	Items       []string            `json:"items"`
	ToProcess   []*NetflixHistory   `json:"-"`
}

func NewHistory() *History {
	return &History{
		ItemsSearch: make(map[string]struct{}),
		ToProcess:   []*NetflixHistory{},
	}
}

func (h *History) Has(item string) bool {
	_, ok := h.ItemsSearch[item]
	return ok
}

func (h *History) Push(item string, r Reporter) {
	if h.Has(item) {
		return
	}

	if len(h.Items) >= NetfliHistorySize {
		delete(h.ItemsSearch, h.Items[0])
		h.Items = h.Items[1:]
	}

	h.Items = append(h.Items, item)
	h.ItemsSearch[item] = struct{}{}
	h.ToProcess = append(h.ToProcess, parseNetflixTitle(item, r))
}

func (h *History) Write() error {
	dataFilePath := filepath.Join(pathutil.ConfigDir(), "history")

	data, err := json.Marshal(h)
	if err != nil {
		return fmt.Errorf("could not Marshal the data: %w", err)
	}
	return os.WriteFile(dataFilePath, data, 0o644)
}

func (h *History) Load() error {
	dataFilePath := filepath.Join(pathutil.ConfigDir(), "history")

	// TODO(melvin): Use something more secure than ReadFile, to avoid
	// loading a huge file in memory.
	data, err := os.ReadFile(dataFilePath)
	if err != nil {
		return fmt.Errorf("could not read the file: %w", err)
	}
	err = json.Unmarshal(data, h)
	if err != nil {
		return fmt.Errorf("could not unmarshal the data: %w", err)
	}
	return nil
}

func (h *History) ClearNetflixHistory() {
	h.ToProcess = []*NetflixHistory{}
}
