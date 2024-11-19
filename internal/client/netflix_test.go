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
	c := New(nil)
	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)
	h, err := c.extractData(bytes.NewReader(data))
	require.NoError(t, err)

	testCases := []struct {
		name    string
		episode string
		isShow  bool
	}{
		{
			name:    "Slasher",
			episode: "Soon Your Own Eyes Will See",
			isShow:  true,
		},
		{
			name:    "That '90s Show",
			episode: "Friends in Low Places",
			isShow:  true,
		},
		{
			name:    "Squid Game: The Challenge",
			episode: "Nowhere To Hide",
			isShow:  true,
		},
		{
			name:    "Alice in Borderland",
			episode: "Episode 8",
			isShow:  true,
		},
		{
			name:    "Strong Girl Nam-soon",
			episode: "Light and Shadow of Gangnam",
			isShow:  true,
		},
		{
			name:    "Goedam",
			episode: "Threshold",
			isShow:  true,
		},
		{
			name:    "Pain Hustlers",
			episode: "",
			isShow:  false,
		},
		{
			name:    "Scott Pilgrim Takes Off",
			episode: "Whatever",
			isShow:  true,
		},
		{
			name:    "Ali Wong: Hard Knock Wife",
			episode: "",
			isShow:  false,
		},
	}
	require.Len(t, h, len(testCases))

	for i, tc := range testCases {
		i := i
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.episode, h[i].EpisodeName)
			assert.Equal(t, tc.name, h[i].Title)
			assert.Equal(t, tc.isShow, h[i].IsShow)
		})
	}
}
