package activitytracker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
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

var wordStartingWithI = regexp.MustCompile(`(?m)(^|[\s\p{P}])i`)

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
			if !stringMatches(r.Movie.Title, h.Title) {
				continue
			}

			medias.Movies = append(medias.Movies, trakt.MarkAsWatched{
				IDs:       r.Movie.IDs,
				WatchedAt: now,
			})
			return nil
		}

		if r.Type == trakt.SearchTypeEpisode {
			if !stringMatches(r.Show.Title, h.Title) || !stringMatches(r.Episode.Title, h.EpisodeName) {
				continue
			}
			if h.Season > 0 && r.Episode.Season != h.Season {
				continue
			}
			medias.Episodes = append(medias.Episodes, trakt.MarkAsWatched{
				IDs:       r.Episode.IDs,
				WatchedAt: now,
			})
			return nil
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
//
// There are a ton of other edge cases we need to account for.
func stringMatches(netflixTitle, traktTitle string) bool {
	// Netflix titles sometimes use "..." to indicate a longer title.
	titleIsPartial := strings.HasSuffix(netflixTitle, "...") && !strings.HasSuffix(traktTitle, "...")

	if areEqual(netflixTitle, traktTitle, titleIsPartial) {
		return true
	}

	stringNormalizer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	netflixTitle, _, err := transform.String(stringNormalizer, netflixTitle)
	if err != nil {
		return false
	}
	traktTitle, _, err = transform.String(stringNormalizer, traktTitle)
	if err != nil {
		return false
	}
	if areEqual(netflixTitle, traktTitle, titleIsPartial) {
		return true
	}

	// Some characters aren't in the trakt title
	charsToReplace := []string{
		// Netflix title: "Arrested Development: Ready, Aim, Marry Me!"
		// Trakt title: "Arrested Development: Ready, Aim, Marry Me"
		"!",
	}

	// Special cases

	// if the title contains "!", then we need to take into account Spanish
	// Ex.
	//   Netflix title: "Arrested Development iAmigos!"
	//   Trakt title: "Arrested Development Amigos"
	//
	// In that example they used an "i" and not a "¡", which is a bit
	// awkward since it forces us to removes all "i"s at the beginning
	// of words.
	if strings.Contains(netflixTitle, "!") || strings.Contains(traktTitle, "!") {
		// We DO NOT remove the 'i's in B to avoid potentially breaking
		// valid titles
		// Ex:
		//   A: iiPhone!
		//   B: iPhone
		// If we cleanup both A and B we would end with
		//   A: iPhone
		//   B: Phone
		netflixTitle = wordStartingWithI.ReplaceAllStringFunc(netflixTitle, func(s string) string {
			// Keep the prefix (space or punctuation), drop the 'i'
			return s[:len(s)-1]
		})
		netflixTitle = strings.ReplaceAll(netflixTitle, "¡", "")
		traktTitle = strings.ReplaceAll(traktTitle, "¡", "")
	}

	for _, char := range charsToReplace {
		netflixTitle = strings.ReplaceAll(netflixTitle, char, "")
		traktTitle = strings.ReplaceAll(traktTitle, char, "")
	}

	// Another edge case we'd rather keep for the end
	// Sometime Netflix titles use spaces instead of dashes:
	// Ex.
	//   Netflix title: "Arrested Development: Forget Me Now"
	//   Trakt title: "Arrested Development: Forget-Me-Now"
	netflixTitle = strings.ReplaceAll(netflixTitle, " ", "-")
	traktTitle = strings.ReplaceAll(traktTitle, " ", "-")

	return areEqual(netflixTitle, traktTitle, titleIsPartial)
}

func areEqual(a, b string, titleIsPartial bool) bool {
	if titleIsPartial {
		// If the title is partial, we need to account for that in our comparison
		if len(a) < 3 {
			return false
		}
		return strings.HasPrefix(b, a[:len(a)-3])
	}
	return strings.EqualFold(a, b)
}
