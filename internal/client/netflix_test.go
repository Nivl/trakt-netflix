package client

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractData(t *testing.T) {
	c := New(nil)
	data, err := os.ReadFile(filepath.Join("testdata", "netflix.html"))
	require.NoError(t, err)
	h, err := c.extractData(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, h, 5)
	//
	require.Equal(t, "Nowhere To Hide", h[0].EpisodeName)
	require.Equal(t, "Squid Game: The Challenge", h[0].Title)
	require.Equal(t, true, h[0].IsShow)
	//
	require.Equal(t, "Light and Shadow of Gangnam", h[1].EpisodeName)
	require.Equal(t, "Strong Girl Nam-soon", h[1].Title)
	require.Equal(t, true, h[1].IsShow)
	//
	require.Equal(t, "", h[2].EpisodeName)
	require.Equal(t, "Pain Hustlers", h[2].Title)
	require.Equal(t, false, h[2].IsShow)
	//
	require.Equal(t, "Whatever", h[3].EpisodeName)
	require.Equal(t, "Scott Pilgrim Takes Off", h[3].Title)
	require.Equal(t, true, h[3].IsShow)
	//
	require.Equal(t, "", h[4].EpisodeName)
	require.Equal(t, "Ali Wong: Hard Knock Wife", h[4].Title)
	require.Equal(t, false, h[4].IsShow)
}
