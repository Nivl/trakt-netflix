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

func (h *NetflixHistory) SearchQuery() string {
	return url.QueryEscape(h.String())
}

var (
	netflixTitleDefaultRegex   = regexp.MustCompile(`(.+): (.+): "(.+)"`)
	netflixTitleShowColonRegex = regexp.MustCompile(`((.+): (.+)): ((.+): (.+)): "(.+)"`)
	netflixTitleSeasonRegex    = regexp.MustCompile(`(.+): ((Season (\d+))|(Limited Series)): "(.+)"`)
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
		// Ex: Alice in Borderland: Season 2: "Episode 8"
		// Ex: Strong Girl Nam-soon: Limited Series: "Forewarned Bloodbath"
		matches = netflixTitleSeasonRegex.FindAllStringSubmatch(title, -1)
		if len(matches) == 1 && len(matches[0]) == 7 {
			h.Title = matches[0][1]
			h.EpisodeName = matches[0][6]
			return
		}

		c.Report("Potentially weird title found: " + title)
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
