package trakt

type SearchTypes string

const (
	SearchTypeMovie   SearchTypes = "movie"
	SearchTypeEpisode SearchTypes = "episode"
)

type IDs struct {
	Trakt int     `json:"trakt"`
	Slug  *string `json:"slug,omitempty"`
	IMDB  *string `json:"imdb,omitempty"`
	TMDB  *int    `json:"tmdb,omitempty"`
	TVDB  *int    `json:"tvdb,omitempty"`
}

type Media struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   IDs    `json:"ids"`
}

type Episode struct {
	Season int    `json:"season"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Year   int    `json:"year"`
	IDs    IDs    `json:"ids"`
}

type MarkAsWatched struct {
	WatchedAt string `json:"watched_at,omitempty"`
	IDs       IDs    `json:"ids"`
}
