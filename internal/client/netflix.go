package client

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// NetflixConfig contains the configuration needed for Netflix
type NetflixConfig struct {
	AccountID string `env:"ACCOUNT_ID"`
	Cookie    string `env:"COOKIE,required"`
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

var netflixTitleRegex = regexp.MustCompile(`(.+): (.+): "(.+)"`)

// FetchNetflixHistory returns the viewing history from Netflix
func (c *Client) FetchNetflixHistory(cfg NetflixConfig) (history []*NetflixHistory, err error) {
	u := "https://www.netflix.com/viewingactivity/" + cfg.AccountID
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

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse HTML: %w", err)
	}

	doc.Find(".retableRow").Each(func(_ int, s *goquery.Selection) {
		title := s.Find(".title").Find("a").Text()
		h := &NetflixHistory{
			Title:  title,
			IsShow: netflixTitleRegex.MatchString(title),
			Date:   s.Find(".date").Text(),
		}
		history = append(history, h)

		if !h.IsShow {
			return
		}

		// Format is `<Show Name>: <Show Name>: "<Episode Name>"``
		matches := netflixTitleRegex.FindAllStringSubmatch(title, -1)
		if (len(matches) != 1 && len(matches[0]) != 4) || matches[0][1] != matches[0][2] {
			slog.Warn("Potentially weird title found", "title", title)
			h.IsShow = false
			return
		}

		h.Title = matches[0][1]
		h.EpisodeName = matches[0][3]
	})

	return history, nil
}
