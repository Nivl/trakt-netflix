package client

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/Nivl/trakt-netflix/internal/trakt"
)

var stringNormalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// MarkAsWatched mark as watched all the provided media
func (c *Client) MarkAsWatched(ctx context.Context) {
	medias := &trakt.MarkAsWatchedRequest{}
	for _, h := range c.history.ToProcess {
		err := c.searchMedia(ctx, h, medias)
		if err != nil {
			c.Report("Trakt: Couldn't find: " + h.String() + ".\nError: " + err.Error() + "\nPlease add manually.")
			slog.Error("media search failed", "isShow", h.IsShow, "media", h.String(), "error", err.Error())
			continue
		}
		c.Report("Adding to current watchlist batch: " + h.String())

		time.Sleep(100 * time.Millisecond)
	}

	_, err := c.traktClient.MarkAsWatched(ctx, medias)
	if err != nil {
		c.Report("Trakt: Couldn't mark the batch as watched. Error: " + err.Error())
		slog.Error("failed to watch", "error", err.Error(), "medias", medias)
		return
	}
	c.Report("Batch processed successfully")
	c.history.ClearNetflixHistory()
}

// searchMedia tries to map a Netflix movie/episode to one on Trakt
func (c *Client) searchMedia(ctx context.Context, h *NetflixHistory, medias *trakt.MarkAsWatchedRequest) error {
	now := time.Now().Format(time.RFC3339)

	typ := trakt.SearchTypeMovie
	if h.IsShow {
		typ = trakt.SearchTypeEpisode
	}

	response, err := c.traktClient.Search(ctx, typ, h.SearchQuery())
	if err != nil {
		return fmt.Errorf("searching for %s: %w", h.SearchQuery(), err)
	}

	for _, r := range response.Results {
		if r.Type == trakt.SearchTypeMovie {
			if stringMatches(r.Movie.Title, h.Title) {
				medias.Movies = append(medias.Movies, trakt.MarkAsWatched{
					IDs:       r.Movie.IDs,
					WatchedAt: now,
				})
				return nil
			}
			continue
		}

		if r.Type == trakt.SearchTypeEpisode {
			if stringMatches(r.Show.Title, h.Title) && stringMatches(r.Episode.Title, h.EpisodeName) {
				medias.Episodes = append(medias.Episodes, trakt.MarkAsWatched{
					IDs:       r.Episode.IDs,
					WatchedAt: now,
				})
				return nil
			}
			continue
		}
	}
	return fmt.Errorf("not found")
}

// Sometime the title don't match due to unicode characters.
// For example,
// On Netflix: "Arrested Development: Beef Consomme"
// On Trakt: "Arrested Development: Beef Consommé"
//
// So on top of regular title search, we also normalize the titles
// to remove accents and diacritics.
//
// Netflix and Trakt may also use different cases for the same title.
// For example,
// On Netflix: "Arrested Development: Justice is Blind"
// On Trakt: "Arrested Development: Justice Is Blind"
func stringMatches(a, b string) bool {
	if strings.EqualFold(a, b) {
		return true
	}

	normalizedA, _, err := transform.String(stringNormalizer, a)
	if err != nil {
		return false
	}
	normalizedB, _, err := transform.String(stringNormalizer, b)
	if err != nil {
		return false
	}
	return strings.EqualFold(normalizedA, normalizedB)
}
