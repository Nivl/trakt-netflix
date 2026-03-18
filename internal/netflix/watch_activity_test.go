package netflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatchActivitySearchTerms(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		activity  WatchActivity
		wantQuery string
		wantShow  string
	}{
		{
			name: "movie",
			activity: WatchActivity{
				Date:        "",
				Title:       "Pain Hustlers",
				EpisodeName: "",
				IsShow:      false,
				Season:      0,
			},
			wantQuery: "Pain Hustlers",
			wantShow:  "",
		},
		{
			name: "episode",
			activity: WatchActivity{
				Date:        "",
				Title:       "Goedam",
				EpisodeName: "Threshold",
				IsShow:      true,
				Season:      0,
			},
			wantQuery: "Threshold",
			wantShow:  "Goedam",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantQuery, tc.activity.SearchQuery())
			assert.Equal(t, tc.wantShow, tc.activity.SearchShow())
		})
	}
}
