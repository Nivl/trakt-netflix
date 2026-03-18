package trakt

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nivl/trakt-netflix/internal/secret"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		req       SearchRequest
		wantQuery map[string]string
	}{
		{
			name: "movie search",
			req: SearchRequest{
				Type:  SearchTypeMovie,
				Query: "Pain Hustlers",
				Show:  "",
			},
			wantQuery: map[string]string{
				"query": "Pain Hustlers",
				"type":  string(SearchTypeMovie),
				"show":  "",
			},
		},
		{
			name: "episode search with show",
			req: SearchRequest{
				Type:  SearchTypeEpisode,
				Query: "Threshold",
				Show:  "Goedam",
			},
			wantQuery: map[string]string{
				"query": "Threshold",
				"type":  string(SearchTypeEpisode),
				"show":  "Goedam",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotPath string
			var gotQuery map[string]string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = map[string]string{
					"query": r.URL.Query().Get("query"),
					"type":  r.URL.Query().Get("type"),
					"show":  r.URL.Query().Get("show"),
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, "[]")
			}))
			t.Cleanup(srv.Close)

			client := new(Client)
			client.http = srv.Client()
			client.baseURL = srv.URL
			client.clientID = "test-client-id"
			client.clientSecret = secret.NewSecret("")

			searchResponse, err := client.Search(t.Context(), tc.req)
			require.NoError(t, err)
			require.NotNil(t, searchResponse)
			assert.Empty(t, searchResponse.Results)
			assert.Equal(t, "/search", gotPath)
			assert.Equal(t, tc.wantQuery, gotQuery)
		})
	}
}
