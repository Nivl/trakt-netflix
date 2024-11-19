package client

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// NetflixConfig contains the configuration needed for Netflix
type NetflixConfig struct {
	AccountID string `env:"ACCOUNT_ID"`
	Cookie    string `env:"COOKIE,required"`
	URL       string `env:"URL,default=https://www.netflix.com/viewingactivity"`
}

// NetflixHistory contains the data from Netflix
type NetflixHistory struct {
	Date        string
	Title       string
	EpisodeName string
	IsShow      bool
}

func (h *NetflixHistory) String() string {
	if h.IsShow {
		return fmt.Sprintf("%s: %s", h.Title, h.EpisodeName)
	}
	return h.Title
}

// SearchQuery returns the query string to use on trakt to search for the media
func (h *NetflixHistory) SearchQuery() string {
	query := h.Title
	if h.IsShow {
		// for some reasons, wrapping the title and the episode name in quotes
		// returns better search results
		query = fmt.Sprintf("%q: %q", h.Title, h.EpisodeName)
	}
	return url.QueryEscape(query)
}

var (
	netflixTitleDefaultRegex   = regexp.MustCompile(`(.+): (.+): "(.+)"`)
	netflixTitleShowColonRegex = regexp.MustCompile(`((.+): (.+)): ((.+): (.+)): "(.+)"`)
	netflixTitleSeasonRegex    = regexp.MustCompile(`(.+): (((Season|Part) (\d+))|(Limited Series)|(Collection)): "(.+)"`)
)

// FetchNetflixHistory returns the viewing history from Netflix
func (c *Client) FetchNetflixHistory(cfg NetflixConfig) (history []*NetflixHistory, err error) {
	slog.Info("Checking for new watched medias on Netflix")
	u := cfg.URL + "/" + cfg.AccountID
	req, err := http.NewRequest(http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "NetflixId",
		Value: cfg.Cookie,
	})

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_, copyErr := io.Copy(io.Discard, res.Body)
		CloseErr := res.Body.Close()
		if err == nil {
			err = copyErr
			if err == nil {
				err = CloseErr
			}
		}
	}()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", res.StatusCode)
	}

	return c.extractData(res.Body)
}

func (c *Client) extractData(r io.Reader) (history []*NetflixHistory, err error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse HTML: %w", err)
	}

	doc.Find(".retableRow").Each(func(_ int, s *goquery.Selection) {
		title := s.Find(".title").Find("a").Text()
		title = cleanupString(title)

		h := &NetflixHistory{
			Title:  title,
			IsShow: netflixTitleDefaultRegex.MatchString(title),
			Date:   s.Find(".date").Text(),
		}
		history = append(history, h)

		if !h.IsShow {
			return
		}

		// Format is `<Show Name>: <Show Name>: "<Episode Name>"`
		// This is the most common format
		matches := netflixTitleDefaultRegex.FindAllStringSubmatch(title, -1)
		if (len(matches) == 1 && len(matches[0]) == 4) && matches[0][1] == matches[0][2] {
			h.Title = matches[0][1]
			h.EpisodeName = matches[0][3]
			return
		}

		// Show with a colon in its name like "Squid Game: The Challenge".
		// Format is `<Show: Name>: <Show: Name>: "<Episode Name>"`
		matches = netflixTitleShowColonRegex.FindAllStringSubmatch(title, -1)
		if (len(matches) == 1 && len(matches[0]) == 8) && matches[0][1] == matches[0][4] {
			h.Title = matches[0][1]
			h.EpisodeName = matches[0][7]
			return
		}

		// Weird edge case: `<Show Name>: Season <number>: "<Episode Name>"`
		//                  `<Show Name>: Limited Series: "<Episode Name>"`
		//                  `<Show Name>: Collection: "<Episode Name>"`
		//                  `<Show Name>: Part <number>: "<Episode Name>"`
		// Ex: Alice in Borderland: Season 2: "Episode 8"
		// Ex: Strong Girl Nam-soon: Limited Series: "Forewarned Bloodbath"
		// Ex: Goedam: Collection: "Birth"
		// Ex: That '90s Show: Part 2: "Friends in Low Places"
		matches = netflixTitleSeasonRegex.FindAllStringSubmatch(title, -1)
		if len(matches) == 1 && len(matches[0]) == 9 {
			h.Title = matches[0][1]
			h.EpisodeName = matches[0][8]
			return
		}

		// Now it gets complicated...
		// Some shows have a subtitle as season name
		// Ex. Slasher: The Executioner: "Soon Your Own Eyes Will See"
		//
		// It's also possible for a movie to have multiple colons in its name.

		// If there's only 2 colons we're going to assume it's a show
		// and we'll drop the middle part
		matches = netflixTitleDefaultRegex.FindAllStringSubmatch(title, -1)
		if len(matches) == 1 && len(matches[0]) == 4 {
			h.Title = matches[0][1]
			h.EpisodeName = matches[0][3]

			c.Report(
				fmt.Sprintf("Potentially weird title found: %s. Assuming it's a show named '%s' with an episode named '%s'",
					title, h.Title, h.EpisodeName,
				))

			return
		}

		c.Report(fmt.Sprintf("Potentially weird title found: %s. Assuming it's a movie.", title))
		h.IsShow = false
	})

	return history, nil
}

func cleanupString(s string) string {
	out := strings.Builder{}
	lastIsSpace := true
	for _, r := range s {
		isSpace := unicode.IsSpace(r)
		if isSpace && lastIsSpace {
			continue
		}
		lastIsSpace = isSpace
		if isSpace {
			out.WriteRune(' ')
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}
