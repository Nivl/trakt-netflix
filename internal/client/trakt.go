package client

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// TraktConfig contains the configuration needed for Netflix
type TraktConfig struct {
	CSRF   string `env:"CSRF,required"`
	Cookie string `env:"COOKIE,required"`
}

// MarkAsWatched mark as watched all the provided media
func (c *Client) MarkAsWatched(conf TraktConfig, history []*NetflixHistory) {
	cfg := &conf
	dataFilePath := filepath.Join(ConfigDir(), "data")
	lastImportedRaw, err := os.ReadFile(dataFilePath)
	if err != nil {
		slog.Warn("could not read last imported file", "error", err.Error())
	}
	lastImported := strings.TrimSpace(strings.Trim(string(lastImportedRaw), "\n"))

	for _, h := range history {
		if h.String() == lastImported {
			break
		}

		u, id, err := c.searchMedia(cfg, h)
		if err != nil {
			c.Report("Trakt: Couldn't find: " + h.String() + ". Error: " + err.Error())
			slog.Error("failed to search", "media", h.String(), "error", err.Error())
			continue
		}

		time.Sleep(500 * time.Millisecond)
		err = c.watch(cfg, h, u, id)
		if err != nil {
			c.Report("Trakt: Couldn't mark as watched: " + h.String() + ". Error: " + err.Error())
			slog.Error("failed to watch", "media", h.String(), "error", err.Error())
			continue
		}

		c.Report("Trakt: Watched " + h.String())
		time.Sleep(500 * time.Millisecond)
	}

	err = os.WriteFile(dataFilePath, []byte(history[0].String()), 0o644)
	if err != nil {
		c.Report(fmt.Sprintf(`Trakt Couldn't update DB with %q. Error: %s`+history[0].String(), err.Error()))
		slog.Error("could not write last imported file", "error", err.Error())
	}
}

// watch builds and sends the http request that marks a media as watched
func (c *Client) watch(cfg *TraktConfig, h *NetflixHistory, u, id string) (err error) {
	mediaType := "movie"
	if h.IsShow {
		mediaType = "episode"
	}
	data := url.Values{}
	data.Set("trakt_id", id)
	data.Set("type", mediaType)
	data.Set("watched_at", "now")
	data.Set("collected_at", "now")
	data.Set("rewatching", "false")
	data.Set("force", "false")

	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "_traktsession",
		Value: cfg.Cookie,
	})
	req.Header.Add("x-csrf-token", cfg.CSRF)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
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

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("request failed with status %d", res.StatusCode)
	}

	return nil
}

// searchMedia tries to map a Netflix movie/episode to one on Trakt
func (c *Client) searchMedia(cfg *TraktConfig, h *NetflixHistory) (watchURL, id string, err error) {
	var u string
	switch h.IsShow {
	case true:
		u = "https://trakt.tv/search/episodes/?query=" + h.SearchQuery()
	default:
		u = "https://trakt.tv/search/movies/?query=" + h.SearchQuery()
	}

	req, err := http.NewRequest(http.MethodGet, u, http.NoBody)
	if err != nil {
		return "", "", fmt.Errorf("could not create request: %w", err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "_traktsession",
		Value: cfg.Cookie,
	})
	req.Header.Add("x-csrf-token", cfg.CSRF)

	res, err := c.http.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("request failed: %w", err)
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
		return "", "", fmt.Errorf("request failed with status %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("couldn't parse HTML: %w", err)
	}

	s := doc.Find(".grid-item").First()
	dataURL, ok := s.Attr("data-url")
	if !ok {
		return "", "", fmt.Errorf("no data-url found")
	}
	dataID, ok := s.Attr("data-episode-id")
	if !ok {
		dataID, ok = s.Attr("data-movie-id")
		if !ok {
			return "", "", fmt.Errorf("no data-episode-id nor data-movie-id")
		}
	}
	return "https://trakt.tv" + dataURL + "/watch", dataID, nil
}
