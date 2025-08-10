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
			title: `Friendly Rivalry: "Episode 16"`,
			expected: &netflix.WatchActivity{
				Title:       "Friendly Rivalry",
				EpisodeName: "Episode 16",
				Season:      1,
				IsShow:      true,
			},
		},
		{
			title: `Zombieverse: New Blood: "Episode 7"`,
			expected: &netflix.WatchActivity{
				Title:       "Zombieverse",
				EpisodeName: "Episode 7",
				Season:      2,
				IsShow:      true,
			},
		},
		{
			title: `The Devil's Plan: Season 2: "Episode 9"`,
			expected: &netflix.WatchActivity{
				Title:       "The Devil's Plan",
				Season:      2,
				EpisodeName: "Episode 9",
				IsShow:      true,
			},
		},
		{
			title: `Squid Game: Season 3: "○△□"`,
			expected: &netflix.WatchActivity{
				Title:       "Squid Game",
				Season:      3,
				EpisodeName: "○△□",
				IsShow:      true,
			},
		},
		{
			title: `Squid Game: Season 3: "Humans Are…"`,
			expected: &netflix.WatchActivity{
				Title:       "Squid Game",
				Season:      3,
				EpisodeName: "Humans Are…",
				IsShow:      true,
			},
		},
		{
			title: `Chicken Nugget: Limited Series: "Episode 5"`,
			expected: &netflix.WatchActivity{
				Title:       "Chicken Nugget",
				EpisodeName: "Episode 5",
				IsShow:      true,
			},
		},
		{
			title: `Old Enough!: Season 2: "Episode 4"`,
			expected: &netflix.WatchActivity{
				Title:       "Old Enough!",
				Season:      2,
				EpisodeName: "Episode 4",
				IsShow:      true,
			},
		},
		{
			title: `Love, Death & Robots: Volume 4: "Close Encounters of the Mini Kind"`,
			expected: &netflix.WatchActivity{
				Title:       "Love, Death & Robots",
				Season:      4,
				EpisodeName: "Close Encounters of the Mini Kind",
				IsShow:      true,
			},
		},
		{
			title: `A Man on the Inside: "The Curious Incident of the Dog in the Painting Class"`,
			expected: &netflix.WatchActivity{
				Title:       "A Man on the Inside",
				EpisodeName: "The Curious Incident of the Dog in the Painting Class",
				IsShow:      true,
				Season:      1,
			},
		},
		{
			title: `Weak Hero: Class 2: "Episode 1"`,
			expected: &netflix.WatchActivity{
				Title:       "Weak Hero",
				Season:      2,
				EpisodeName: "Episode 1",
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
			title: `Arrested Development: Season 4 Remix: Fateful Consequences: "A Couple-A New Starts"`,
			expected: &netflix.WatchActivity{
				Title:       "Arrested Development",
				Season:      4,
				EpisodeName: "A Couple-A New Starts",
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
