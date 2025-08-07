package netflix_test

import (
	"testing"

	"github.com/Nivl/trakt-netflix/internal/netflix"
	"github.com/stretchr/testify/assert"
)

func TestParseTitle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title    string
		expected *netflix.WatchActivity
	}{
		{
			title: `Arrested Development: Season 1: "Justice is Blind"`,
			expected: &netflix.WatchActivity{
				Title:       "Arrested Development",
				EpisodeName: "Justice is Blind",
				Season:      1,
				IsShow:      true,
			},
		},
		{
			title: `Goedam: Collection: "Threshold"`,
			expected: &netflix.WatchActivity{
				Title:       "Goedam",
				EpisodeName: "Threshold",
				IsShow:      true,
			},
		},
		{
			title: `Scott Pilgrim Takes Off: Scott Pilgrim Takes Off: "Whatever"`,
			expected: &netflix.WatchActivity{
				Title:       "Scott Pilgrim Takes Off",
				EpisodeName: "Whatever",
				IsShow:      true,
			},
		},
		{
			title: `Strong Girl Nam-soon: Limited Series: "Light and Shadow of Gangnam"`,
			expected: &netflix.WatchActivity{
				Title:       "Strong Girl Nam-soon",
				EpisodeName: "Light and Shadow of Gangnam",
				IsShow:      true,
			},
		},
		{
			title: `Alice in Borderland: Season 2: "Episode 8"`,
			expected: &netflix.WatchActivity{
				Title:       "Alice in Borderland",
				Season:      2,
				EpisodeName: "Episode 8",
				IsShow:      true,
			},
		},
		{
			title: `Squid Game: The Challenge: Squid Game: The Challenge: "Nowhere To Hide"`,
			expected: &netflix.WatchActivity{
				Title:       "Squid Game: The Challenge",
				EpisodeName: "Nowhere To Hide",
				IsShow:      true,
			},
		},
		{
			title: `That '90s Show: Part 2: "Friends in Low Places"`,
			expected: &netflix.WatchActivity{
				Title:       "That '90s Show",
				Season:      2,
				EpisodeName: "Friends in Low Places",
				IsShow:      true,
			},
		},
		{
			title: `Slasher: The Executioner: "Soon Your Own Eyes Will See"`,
			expected: &netflix.WatchActivity{
				Title:       "Slasher",
				EpisodeName: "Soon Your Own Eyes Will See",
				IsShow:      true,
			},
		},
		{
			title: `Pain Hustlers`,
			expected: &netflix.WatchActivity{
				Title:  "Pain Hustlers",
				IsShow: false,
			},
		},
		{
			title: `Ali Wong: Hard Knock Wife`,
			expected: &netflix.WatchActivity{
				Title:  "Ali Wong: Hard Knock Wife",
				IsShow: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			res := netflix.ParseTitle(t.Context(), tc.title, nil)
			assert.Equal(t, tc.expected, res)
		})
	}
}
