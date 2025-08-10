package netflix

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Nivl/trakt-netflix/internal/o11y"
)

var (
	titleDefaultRegex   = regexp.MustCompile(`(.+): (.+): "(.+)"`)
	titleShowColonRegex = regexp.MustCompile(`((.+): (.+)): ((.+): (.+)): "(.+)"`)
	titleSeasonRegex    = regexp.MustCompile(`(.+): (((Season|Part|Class|Volume) (\d+)((.+)?: .+)?)|(Limited Series)|(Collection)): "(.+)"`)
	titleShortRegex     = regexp.MustCompile(`([^:]+): "([^:]+)"`)
)

// ParseTitle parses a Netflix title and turns it into a WatchActivity.
func ParseTitle(ctx context.Context, title string, reporter o11y.Reporter) *WatchActivity {
	h := &WatchActivity{ //nolint:exhaustruct // The point of this function is to slowly build that object
		Title: title,
		// All shows have their episode names wrapped in quotes.
		// It doesn't mean that *only* shows have quotes, but it's a good
		// way to exit early and prevent too many shenanigans with parsing
		// movie titles. Very few movies have quotes in their title.
		IsShow: strings.Contains(title, `"`),
	}

	if !h.IsShow {
		return h
	}

	// Format is `<Show Name>: <Show Name>: "<Episode Name>"`
	// This is the most common format
	matches := titleDefaultRegex.FindAllStringSubmatch(title, -1)
	if (len(matches) == 1 && len(matches[0]) == 4) && matches[0][1] == matches[0][2] {
		h.Title = matches[0][1]
		h.EpisodeName = matches[0][3]
		return h
	}

	// Show with a colon in its name like "Squid Game: The Challenge".
	// Format is `<Show: Name>: <Show: Name>: "<Episode Name>"`
	matches = titleShowColonRegex.FindAllStringSubmatch(title, -1)
	if (len(matches) == 1 && len(matches[0]) == 8) && matches[0][1] == matches[0][4] {
		h.Title = matches[0][1]
		h.EpisodeName = matches[0][7]
		return h
	}

	// Weird edge case: `<Show Name>: Season <number>: "<Episode Name>"`
	//                  `<Show Name>: Limited Series: "<Episode Name>"`
	//                  `<Show Name>: Collection: "<Episode Name>"`
	//                  `<Show Name>: Part <number>: "<Episode Name>"`
	//                  `<Show Name>: Class <number>: "<Episode Name>"`
	//                  `<Show Name>: Volume <number>: "<Episode Name>"`
	//                  `<Show Name>: Season <number> <extra words>: <Season name>: "<Episode Name>"`
	// Ex: Alice in Borderland: Season 2: "Episode 8"
	// Ex: Strong Girl Nam-soon: Limited Series: "Forewarned Bloodbath"
	// Ex: Goedam: Collection: "Birth"
	// Ex: That '90s Show: Part 2: "Friends in Low Places"
	// Ex: Weak Hero: Class 2: "Episode 1"
	// Ex: Love, Death & Robots: Volume 4: "Close Encounters of the Mini Kind"
	// Ex: Arrested Development: Season 4 Remix: Fateful Consequences: "A Couple-A New Starts"
	matches = titleSeasonRegex.FindAllStringSubmatch(title, -1)
	if len(matches) == 1 && len(matches[0]) == 11 {
		h.Title = matches[0][1]
		h.EpisodeName = matches[0][10]
		// It's expected that it may fail if there is no season number
		h.Season, _ = strconv.Atoi(matches[0][5])
		return h
	}

	// Hardcoded cases for things we cannot figure out programatically
	// without accessing Netflix's private graphQL API
	//
	// Zombieverse season 1 has a regular pattern, but for season 2
	// they decided to drop the season number and put a season name instead.
	// That would be fine if the episode also had names. Since they don't,
	// we have to hardcode the mapping. because otherwise we don't
	// know what's season 1, and what's season 2.
	//
	// Format is:`Zombieverse: New Blood: "Episode <number>"`
	prefix := "Zombieverse: New Blood: "
	if strings.HasPrefix(title, prefix) {
		h.Season = 2
		h.Title = "Zombieverse"
		h.EpisodeName = strings.Trim(strings.TrimPrefix(title, prefix), "\"")
		return h
	}

	// Now it gets complicated...
	// Some shows have a subtitle as season name
	// Ex. Slasher: The Executioner: "Soon Your Own Eyes Will See"
	//
	// It's also possible for a movie to have multiple colons in its name.

	// If there's only 2 colons we're going to assume it's a show
	// and we'll drop the middle part
	matches = titleDefaultRegex.FindAllStringSubmatch(title, -1)
	if len(matches) == 1 && len(matches[0]) == 4 {
		h.Title = matches[0][1]
		h.EpisodeName = matches[0][3]

		if reporter != nil {
			reporter.SendMessage(ctx,
				fmt.Sprintf("Potentially weird title found: %s. Assuming it's a show named '%s' with an episode named '%s'",
					title, h.Title, h.EpisodeName,
				))
		}

		return h
	}

	// Last but not least, some shows don't have season markers and don't
	// have their names repeated twice.
	// This is annoying as hell because it can match movies too.
	//
	// Format is: `<Show Name>: "<Episode Name>"`
	matches = titleShortRegex.FindAllStringSubmatch(title, -1)
	if len(matches) == 1 && len(matches[0]) == 3 {
		h.Title = matches[0][1]
		h.EpisodeName = matches[0][2]
		h.Season = 1
		return h
	}

	if reporter != nil {
		reporter.SendMessage(ctx, fmt.Sprintf("Potentially weird title found: %s. Assuming it's a movie.", title))
	}
	h.IsShow = false
	return h
}
