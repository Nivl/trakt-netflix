package activitytracker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/Nivl/trakt-netflix/internal/netflix"
	"github.com/Nivl/trakt-netflix/internal/slack"
	"github.com/Nivl/trakt-netflix/internal/trakt"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var stringNormalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// Client represents a client to interact with external services
type Client struct {
	traktClient   *trakt.Client
	netflixClient *netflix.Client
	slackClient   *slack.Client
}

// New returns a new Client
func New(traktClient *trakt.Client, netflixClient *netflix.Client, slackClient *slack.Client) *Client {
	return &Client{
		slackClient:   slackClient,
		traktClient:   traktClient,
		netflixClient: netflixClient,
	}
}

// Run fetches the viewing history from Netflix and marks it as
// watched on Trakt
func (c *Client) Run(ctx context.Context) error {
	if err := c.UpdateHistory(ctx); err != nil {
		return err
	}
	c.MarkAsWatched(ctx)
	if err := c.netflixClient.History.Write(); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

// UpdateHistory fetches the viewing history from Netflix and
// updates the local history.
func (c *Client) UpdateHistory(ctx context.Context) error {
	err := c.netflixClient.UpdateHistory(ctx, c.slackClient)
	if err != nil {
		return fmt.Errorf("update history: %w", err)
	}
	return nil
}

// MarkAsWatched mark as watched all the provided media
func (c *Client) MarkAsWatched(ctx context.Context) {
	medias := new(trakt.MarkAsWatchedRequest)
	for _, h := range c.netflixClient.History.NewActivity {
		err := c.searchMedia(ctx, h, medias)
		if err != nil {
			c.slackClient.SendMessage(ctx, "Trakt: Couldn't find: "+h.String()+"\nError: "+err.Error()+"\nPlease add manually.")
			slog.ErrorContext(ctx, "media search failed", "isShow", h.IsShow, "media", h.String(), "error", err.Error())
			continue
		}
		c.slackClient.SendMessage(ctx, "Adding to current watchlist batch: "+h.String())

		time.Sleep(100 * time.Millisecond)
	}

	_, err := c.traktClient.MarkAsWatched(ctx, medias)
	if err != nil {
		c.slackClient.SendMessage(ctx, "Trakt: Couldn't mark the batch as watched. Error: "+err.Error())
		slog.ErrorContext(ctx, "failed to watch", "error", err.Error(), "medias", medias)
		return
	}
	c.slackClient.SendMessage(ctx, "Batch processed successfully")
	c.netflixClient.History.ClearNewActivity()
}

// searchMedia tries to map a Netflix movie/episode to one on Trakt
func (c *Client) searchMedia(ctx context.Context, h *netflix.WatchActivity, medias *trakt.MarkAsWatchedRequest) error {
	now := time.Now().Format(time.RFC3339)

	typ := trakt.SearchTypeMovie
	if h.IsShow {
		typ = trakt.SearchTypeEpisode
	}

	response, err := c.traktClient.Search(ctx, typ, h.SearchQuery())
	if err != nil {
		return fmt.Errorf("searching for %s: %w", h.SearchQuery(), err)
	}

	for i := range response.Results {
		r := &response.Results[i]
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
	return errors.New("not found")
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
	if strings.EqualFold(normalizedA, normalizedB) {
		return true
	}

	// Some character aren't in the trakt title
	charsToReplace := []string{
		"!", // Arrested Development Ready, Aim, Marry Me!
	}

	// Special cases

	// if the title contains "!", then we need to take into account Spanish
	// Ex. Arrested Development "iAmigos!"
	// In that example they used a "i" and not a "¡", which makes
	// everything a bit awkward since it forces us to remove all "i"s.
	if strings.Contains(normalizedA, "!") || strings.Contains(normalizedB, "!") {
		normalizedA = strings.ReplaceAll(normalizedA, "i", "")
		normalizedB = strings.ReplaceAll(normalizedB, "i", "")

		normalizedA = strings.ReplaceAll(normalizedA, "¡", "")
		normalizedB = strings.ReplaceAll(normalizedB, "¡", "")
	}

	for _, char := range charsToReplace {
		normalizedA = strings.ReplaceAll(normalizedA, char, "")
		normalizedB = strings.ReplaceAll(normalizedB, char, "")
	}

	return strings.EqualFold(normalizedA, normalizedB)
}
