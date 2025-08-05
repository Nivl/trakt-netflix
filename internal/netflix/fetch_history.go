package netflix

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"unicode"

	"github.com/Nivl/trakt-netflix/internal/errutil"
	"github.com/PuerkitoBio/goquery"
)

// FetchHistory returns the viewing history from Netflix
//
// The returned data must be closed
func (c *Client) FetchHistory(ctx context.Context) (history []string, err error) {
	slog.Info("Checking for new watched medias on Netflix")

	res, err := c.request(ctx, c.watchActivityURL)
	if err != nil {
		return nil, fmt.Errorf("make http request: %w", err)
	}

	defer errutil.RunAndSetError(res.Body.Close, &err, "close response body")
	defer errutil.RunAndSetError(func() error {
		_, err = io.Copy(io.Discard, res.Body)
		return err
	}, &err, "empty response body")

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", res.StatusCode)
	}

	return c.extractData(res.Body)
}

func (c *Client) extractData(r io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	newList := make([]string, 0, HistorySize)
	for _, s := range doc.Find(".retableRow").EachIter() {
		newList = append(newList, s.Find(".title").Find("a").Text())
	}

	// we reverse the list to have the oldest entries first, and
	// newest last
	slices.Reverse(newList)
	for i, title := range newList {
		newList[i] = cleanupString(title)
	}
	return newList, nil
}

// cleanupString normalizes whitespace in a string.
// TODO(melvin): There's probably a cleaner way to do that.
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
