package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type History struct {
	ItemsSearch map[string]struct{} `json:"search"`
	Items       []string            `json:"items"`
}

func NewHistory() *History {
	return &History{
		ItemsSearch: make(map[string]struct{}),
		Items:       make([]string, 0, 20),
	}
}

func (h *History) Has(item string) bool {
	_, ok := h.ItemsSearch[item]
	return ok
}

func (h *History) Push(item string) {
	if h.Has(item) {
		return
	}

	if len(h.Items) >= 20 {
		delete(h.ItemsSearch, h.Items[0])
		h.Items = h.Items[1:]
	}

	h.Items = append(h.Items, item)
	h.ItemsSearch[item] = struct{}{}
}

func (h *History) Write() error {
	dataFilePath := filepath.Join(ConfigDir(), "history")

	data, err := json.Marshal(h)
	if err != nil {
		return fmt.Errorf("could not Marshal the data", err)
	}
	return os.WriteFile(dataFilePath, data, 0o644)
}

func (h *History) Load() error {
	dataFilePath := filepath.Join(ConfigDir(), "history")
	data, err := os.ReadFile(dataFilePath)
	if err != nil {
		return fmt.Errorf("could not read the file", err)
	}
	return json.Unmarshal(data, h)
}
