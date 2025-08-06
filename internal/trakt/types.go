package trakt

// SearchTypes represents the different types of content that can be
// searched.
type SearchTypes string

const (
	// SearchTypeMovie represents a movie search.
	SearchTypeMovie SearchTypes = "movie"
	// SearchTypeEpisode represents an episode search.
	SearchTypeEpisode SearchTypes = "episode"
)

// IDs represents the various IDs associated with a movie, show,
// actor/actress, etc.
type IDs struct {
	Trakt int     `json:"trakt"`
	Slug  *string `json:"slug,omitempty"`
	IMDB  *string `json:"imdb,omitempty"`
	TMDB  *int    `json:"tmdb,omitempty"`
	TVDB  *int    `json:"tvdb,omitempty"`
}

// Media represents a media item (movie, or show, etc.) in the Trakt API.
type Media struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   IDs    `json:"ids"`
}

// Episode represents a TV episode in the Trakt API.
type Episode struct {
	Season int    `json:"season"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Year   int    `json:"year"`
	IDs    IDs    `json:"ids"`
}

// MarkAsWatched represents a watched item.
type MarkAsWatched struct {
	WatchedAt string `json:"watched_at,omitempty"`
	IDs       IDs    `json:"ids"`
}
