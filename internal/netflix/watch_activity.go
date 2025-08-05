package netflix

import (
	"fmt"
	"net/url"
)

// WatchActivity contains the data from Netflix
type WatchActivity struct {
	Date        string
	Title       string
	EpisodeName string
	IsShow      bool
}

func (h *WatchActivity) String() string {
	if h.IsShow {
		return fmt.Sprintf("%s: %s", h.Title, h.EpisodeName)
	}
	return h.Title
}

// SearchQuery returns the query string to use on trakt to search for the media
func (h *WatchActivity) SearchQuery() string {
	query := h.Title
	if h.IsShow {
		query = fmt.Sprintf("%s %s", h.Title, h.EpisodeName)
	}
	return url.QueryEscape(query)
}
