package netflix

import (
	"fmt"
	"regexp"

	"github.com/Nivl/trakt-netflix/internal/o11y"
)

var (
	titleDefaultRegex   = regexp.MustCompile(`(.+): (.+): "(.+)"`)
	titleShowColonRegex = regexp.MustCompile(`((.+): (.+)): ((.+): (.+)): "(.+)"`)
	titleSeasonRegex    = regexp.MustCompile(`(.+): (((Season|Part) (\d+))|(Limited Series)|(Collection)): "(.+)"`)
)

func ParseTitle(title string, repporter o11y.Reporter) *WatchActivity {
	h := &WatchActivity{
		Title:  title,
		IsShow: titleDefaultRegex.MatchString(title),
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
	// Ex: Alice in Borderland: Season 2: "Episode 8"
	// Ex: Strong Girl Nam-soon: Limited Series: "Forewarned Bloodbath"
	// Ex: Goedam: Collection: "Birth"
	// Ex: That '90s Show: Part 2: "Friends in Low Places"
	matches = titleSeasonRegex.FindAllStringSubmatch(title, -1)
	if len(matches) == 1 && len(matches[0]) == 9 {
		h.Title = matches[0][1]
		h.EpisodeName = matches[0][8]
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

		if repporter != nil {
			repporter.SendMessage(
				fmt.Sprintf("Potentially weird title found: %s. Assuming it's a show named '%s' with an episode named '%s'",
					title, h.Title, h.EpisodeName,
				))
		}

		return h
	}

	if repporter != nil {
		repporter.SendMessage(fmt.Sprintf("Potentially weird title found: %s. Assuming it's a movie.", title))
	}
	h.IsShow = false
	return h
}
