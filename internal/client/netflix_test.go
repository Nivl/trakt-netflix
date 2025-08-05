package client

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

func TestExtractData(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)

	mockctrl := gomock.NewController(t)
	defer mockctrl.Finish()

	Doer := mocks.NewMockDoer(mockctrl)
	Doer.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})

	netflixClient, err := netflix.NewClientWithDoer(netflix.Config{
		URL:       "https://www.netflix.com/viewingactivity",
		AccountID: "your_account_id",
		Cookie:    "your_cookie",
	}, Doer)
	require.NoError(t, err)

	traktClient, err := trakt.NewClient(trakt.ClientConfig{})
	require.NoError(t, err)

	c := New(nil, NewHistory(), traktClient, netflixClient)
	require.NoError(t, err)

	err = c.FetchHistory(t.Context())
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
	show1 := `Scott Pilgrim Takes Off: Scott Pilgrim Takes Off: "Whatever"`
	show2 := `Ali Wong: Hard Knock Wife`
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

	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)

	mockctrl := gomock.NewController(t)
	defer mockctrl.Finish()

	Doer := mocks.NewMockDoer(mockctrl)
	Doer.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})

	netflixClient, err := netflix.NewClientWithDoer(netflix.Config{
		URL:       "https://www.netflix.com/viewingactivity",
		AccountID: "your_account_id",
		Cookie:    "your_cookie",
	}, Doer)
	require.NoError(t, err)

	traktClient, err := trakt.NewClient(trakt.ClientConfig{})
	require.NoError(t, err)

	c := New(nil, history, traktClient, netflixClient)
	require.NoError(t, err)

	err = c.FetchHistory(t.Context())
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
