package netflix

import (
	"fmt"
)

// WatchActivity contains the data from Netflix
type WatchActivity struct {
	Date        string
	Title       string
	EpisodeName string
	IsShow      bool
	Season      int
}

// String implements the Stringer interface.
// it returns a string representation of a WatchActivity.
func (h *WatchActivity) String() string {
	if h.IsShow {
		return fmt.Sprintf("%s: %s", h.Title, h.EpisodeName)
	}
	return h.Title
}

// SearchQuery returns the title to search for on Trakt.
func (h *WatchActivity) SearchQuery() string {
	if h.IsShow {
		return h.EpisodeName
	}
	return h.Title
}

// SearchShow returns the show title when the activity is for an episode search.
func (h *WatchActivity) SearchShow() string {
	if !h.IsShow {
		return ""
	}
	return h.Title
}
