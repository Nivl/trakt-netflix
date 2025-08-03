package client

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractData(t *testing.T) {
	c := New(nil, NewHistory(), nil)
	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)
	err = c.extractData(bytes.NewReader(data))
	require.NoError(t, err)

	testCases := []struct {
		entry   string
		name    string
		episode string
		isShow  bool
	}{
		{
			entry:   `Ali Wong: Hard Knock Wife`,
			name:    "Ali Wong: Hard Knock Wife",
			episode: "",
			isShow:  false,
		},
		{
			entry:   `Scott Pilgrim Takes Off: Scott Pilgrim Takes Off: "Whatever"`,
			name:    "Scott Pilgrim Takes Off",
			episode: "Whatever",
			isShow:  true,
		},
		{
			entry:   `Pain Hustlers`,
			name:    "Pain Hustlers",
			episode: "",
			isShow:  false,
		},
		{
			entry:   `Goedam: Collection: "Threshold"`,
			name:    "Goedam",
			episode: "Threshold",
			isShow:  true,
		},
		{
			entry:   `Strong Girl Nam-soon: Limited Series: "Light and Shadow of Gangnam"`,
			name:    "Strong Girl Nam-soon",
			episode: "Light and Shadow of Gangnam",
			isShow:  true,
		},
		{
			entry:   `Alice in Borderland: Season 2: "Episode 8"`,
			name:    "Alice in Borderland",
			episode: "Episode 8",
			isShow:  true,
		},
		{
			entry:   `Squid Game: The Challenge: Squid Game: The Challenge: "Nowhere To Hide"`,
			name:    "Squid Game: The Challenge",
			episode: "Nowhere To Hide",
			isShow:  true,
		},
		{
			entry:   `That '90s Show: Part 2: "Friends in Low Places"`,
			name:    "That '90s Show",
			episode: "Friends in Low Places",
			isShow:  true,
		},
		{
			entry:   `Slasher: The Executioner: "Soon Your Own Eyes Will See"`,
			name:    "Slasher",
			episode: "Soon Your Own Eyes Will See",
			isShow:  true,
		},
	}
	require.Len(t, c.history.ToProcess, len(testCases))

	for i, tc := range testCases {
		i := i
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := c.history.ToProcess
			assert.Equal(t, tc.episode, h[i].EpisodeName)
			assert.Equal(t, tc.name, h[i].Title)
			assert.Equal(t, tc.isShow, h[i].IsShow)

			item := c.history.Items
			assert.Equal(t, tc.entry, item[i])
		})
	}
}

func TestExtractDataWithExistingData(t *testing.T) {
	show1 := cleanupString(`Scott Pilgrim Takes Off: Scott Pilgrim Takes Off: "Whatever"`)
	show2 := cleanupString(`Ali Wong: Hard Knock Wife`)
	history := &History{
		Items: []string{
			show1,
			show2,
		},
		ItemsSearch: map[string]struct{}{
			show1: {},
			show2: {},
		},
	}
	c := New(nil, history, nil)
	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)
	err = c.extractData(bytes.NewReader(data))
	require.NoError(t, err)

	testCases := []struct {
		entry   string
		name    string
		episode string
		isShow  bool
	}{
		{
			entry:   `Pain Hustlers`,
			name:    "Pain Hustlers",
			episode: "",
			isShow:  false,
		},
		{
			entry:   `Goedam: Collection: "Threshold"`,
			name:    "Goedam",
			episode: "Threshold",
			isShow:  true,
		},
		{
			entry:   `Strong Girl Nam-soon: Limited Series: "Light and Shadow of Gangnam"`,
			name:    "Strong Girl Nam-soon",
			episode: "Light and Shadow of Gangnam",
			isShow:  true,
		},
		{
			entry:   `Alice in Borderland: Season 2: "Episode 8"`,
			name:    "Alice in Borderland",
			episode: "Episode 8",
			isShow:  true,
		},
		{
			entry:   `Squid Game: The Challenge: Squid Game: The Challenge: "Nowhere To Hide"`,
			name:    "Squid Game: The Challenge",
			episode: "Nowhere To Hide",
			isShow:  true,
		},
		{
			entry:   `That '90s Show: Part 2: "Friends in Low Places"`,
			name:    "That '90s Show",
			episode: "Friends in Low Places",
			isShow:  true,
		},
		{
			entry:   `Slasher: The Executioner: "Soon Your Own Eyes Will See"`,
			name:    "Slasher",
			episode: "Soon Your Own Eyes Will See",
			isShow:  true,
		},
	}
	require.Len(t, c.history.ToProcess, len(testCases))

	for i, tc := range testCases {
		i := i
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := c.history.ToProcess
			assert.Equal(t, tc.episode, h[i].EpisodeName)
			assert.Equal(t, tc.name, h[i].Title)
			assert.Equal(t, tc.isShow, h[i].IsShow)
		})
	}
}
