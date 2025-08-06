package activitytracker

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nivl/trakt-netflix/internal/mocks"
	"github.com/Nivl/trakt-netflix/internal/netflix"
	"github.com/Nivl/trakt-netflix/internal/trakt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFetchHistory(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)

	mockctrl := gomock.NewController(t)
	t.Cleanup(mockctrl.Finish)

	Doer := mocks.NewMockDoer(mockctrl)
	Doer.EXPECT().Do(gomock.Any()).DoAndReturn(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})

	netflixClient := &netflix.Client{
		History: &netflix.History{
			ItemsSearch: make(map[string]struct{}),
			Items:       []string{},
			NewActivity: []*netflix.WatchActivity{},
		},
		Cookie:           "cookie",
		WatchActivityURL: "https://www.netflix.com/viewingactivity",
		HTTP:             Doer,
	}

	var traktCfg trakt.ClientConfig
	traktClient, err := trakt.NewClient(traktCfg)
	require.NoError(t, err)

	c := New(traktClient, netflixClient, nil)
	require.NoError(t, err)

	err = c.UpdateHistory(t.Context())
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
	history := c.netflixClient.History
	require.Len(t, history.NewActivity, len(testCases))

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := history.NewActivity
			assert.Equal(t, tc.episode, h[i].EpisodeName)
			assert.Equal(t, tc.name, h[i].Title)
			assert.Equal(t, tc.isShow, h[i].IsShow)

			item := history.Items
			assert.Equal(t, tc.entry, item[i])
		})
	}
}

func TestFetchHistoryWithExistingData(t *testing.T) {
	t.Parallel()

	show1 := `Scott Pilgrim Takes Off: Scott Pilgrim Takes Off: "Whatever"`
	show2 := `Ali Wong: Hard Knock Wife`
	history := &netflix.History{
		Items: []string{
			show1,
			show2,
		},
		ItemsSearch: map[string]struct{}{
			show1: {},
			show2: {},
		},
		NewActivity: []*netflix.WatchActivity{},
	}

	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)

	mockctrl := gomock.NewController(t)
	t.Cleanup(mockctrl.Finish)

	Doer := mocks.NewMockDoer(mockctrl)
	Doer.EXPECT().Do(gomock.Any()).DoAndReturn(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})

	netflixClient := &netflix.Client{
		History:          history,
		Cookie:           "cookie",
		WatchActivityURL: "https://www.netflix.com/viewingactivity",
		HTTP:             Doer,
	}

	var traktCfg trakt.ClientConfig
	traktClient, err := trakt.NewClient(traktCfg)
	require.NoError(t, err)

	c := New(traktClient, netflixClient, nil)
	require.NoError(t, err)

	err = c.UpdateHistory(t.Context())
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
	require.Len(t, history.NewActivity, len(testCases))

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := history.NewActivity
			assert.Equal(t, tc.episode, h[i].EpisodeName)
			assert.Equal(t, tc.name, h[i].Title)
			assert.Equal(t, tc.isShow, h[i].IsShow)
		})
	}
}
