package trakt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Nivl/trakt-netflix/internal/secret"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestSearch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		req             SearchRequest
		wantPath        string
		wantQuery       map[string]string
		wantShowPresent bool
	}{
		{
			name: "movie search",
			req: SearchRequest{
				Type:  SearchTypeMovie,
				Query: "Pain Hustlers",
				Show:  "",
			},
			wantPath: "/search/movie",
			wantQuery: map[string]string{
				"query": "Pain Hustlers",
			},
			wantShowPresent: false,
		},
		{
			name: "episode search with show",
			req: SearchRequest{
				Type:  SearchTypeEpisode,
				Query: "Threshold",
				Show:  "Goedam",
			},
			wantPath: "/search/episode",
			wantQuery: map[string]string{
				"query": "Threshold",
			},
			wantShowPresent: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotPath string
			var gotQuery map[string]string
			var gotShowPresent bool
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				q := r.URL.Query()
				gotQuery = map[string]string{
					"query": q.Get("query"),
				}
				gotShowPresent = q.Has("show")

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
			assert.Equal(t, tc.wantPath, gotPath)
			assert.Equal(t, tc.wantQuery, gotQuery)
			assert.Equal(t, tc.wantShowPresent, gotShowPresent)
		})
	}
}

func TestGetShowSeasons(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		showID       string
		withEpisodes bool
		wantQuery    string
		responseBody string
	}{
		{
			name:         "without episodes",
			showID:       "search-party",
			withEpisodes: false,
			wantQuery:    "",
			responseBody: `[{"number":1,"ids":{"trakt":101}}]`,
		},
		{
			name:         "with episodes",
			showID:       "search-party",
			withEpisodes: true,
			wantQuery:    "episodes",
			responseBody: `[{"number":1,"ids":{"trakt":101},"episodes":[{"season":1,"number":1,"title":"Episode 1","ids":{"trakt":1001}}]}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotPath string
			var gotExtended string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotExtended = r.URL.Query().Get("extended")

				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, tc.responseBody)
			}))
			t.Cleanup(srv.Close)

			client := new(Client)
			client.http = srv.Client()
			client.baseURL = srv.URL
			client.clientID = "test-client-id"
			client.clientSecret = secret.NewSecret("")

			seasons, err := client.GetShowSeasons(t.Context(), tc.showID, tc.withEpisodes)
			require.NoError(t, err)
			require.NotEmpty(t, seasons)
			assert.Equal(t, "/shows/"+tc.showID+"/seasons", gotPath)
			assert.Equal(t, tc.wantQuery, gotExtended)
		})
	}
}

func TestGetSeasonEpisodes(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"season":2,"number":1,"title":"Episode 1","ids":{"trakt":2001}}]`)
	}))
	t.Cleanup(srv.Close)

	client := new(Client)
	client.http = srv.Client()
	client.baseURL = srv.URL
	client.clientID = "test-client-id"
	client.clientSecret = secret.NewSecret("")

	episodes, err := client.GetSeasonEpisodes(t.Context(), "search-party", 2)
	require.NoError(t, err)
	require.Len(t, episodes, 1)
	assert.Equal(t, "/shows/search-party/seasons/2", gotPath)
	assert.Equal(t, "Episode 1", episodes[0].Title)
}

func TestSearchRetriesTransientTransportTimeouts(t *testing.T) {
	t.Parallel()

	attempts := 0
	client := new(Client)
	client.http = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 3 {
				return nil, &url.Error{
					Op:  req.Method,
					URL: req.URL.String(),
					Err: context.DeadlineExceeded,
				}
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("[]")),
				Request:    req,
			}, nil
		}),
	}
	client.baseURL = "https://example.test"
	client.clientID = "test-client-id"
	client.clientSecret = secret.NewSecret("")
	client.retrySleep = func(time.Duration) {}

	searchResponse, err := client.Search(t.Context(), SearchRequest{
		Type:  SearchTypeMovie,
		Query: "Pain Hustlers",
		Show:  "",
	})
	require.NoError(t, err)
	require.NotNil(t, searchResponse)
	assert.Empty(t, searchResponse.Results)
	assert.Equal(t, 3, attempts)
}

func TestGenerateAuthCodeRetriesTransientTransportTimeoutsAndReplaysBody(t *testing.T) {
	t.Parallel()

	attempts := 0
	var bodies []string
	client := new(Client)
	client.http = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			payload, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			bodies = append(bodies, string(payload))

			attempts++
			if attempts == 1 {
				return nil, &url.Error{
					Op:  req.Method,
					URL: req.URL.String(),
					Err: context.DeadlineExceeded,
				}
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"device_code":"device","user_code":"user","verification_url":"https://trakt.tv/activate","expires_in":600,"interval":5}`)),
				Request:    req,
			}, nil
		}),
	}
	client.baseURL = "https://example.test"
	client.clientID = "test-client-id"
	client.clientSecret = secret.NewSecret("")
	client.retrySleep = func(time.Duration) {}

	resp, err := client.GenerateAuthCode(t.Context())
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, attempts)
	assert.Equal(t, []string{`{"client_id":"test-client-id"}`, `{"client_id":"test-client-id"}`}, bodies)
}

func TestSearchDoesNotRetryWhenServerResponds(t *testing.T) {
	t.Parallel()

	attempts := 0
	client := new(Client)
	client.http = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			return &http.Response{
				StatusCode: http.StatusGatewayTimeout,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"error":"timeout"}`)),
				Request:    req,
			}, nil
		}),
	}
	client.baseURL = "https://example.test"
	client.clientID = "test-client-id"
	client.clientSecret = secret.NewSecret("")
	client.retrySleep = func(time.Duration) {}

	searchResponse, err := client.Search(t.Context(), SearchRequest{
		Type:  SearchTypeMovie,
		Query: "Pain Hustlers",
		Show:  "",
	})
	require.Nil(t, searchResponse)
	require.Error(t, err)
	assert.Equal(t, 1, attempts)
	assert.ErrorContains(t, err, "http 504")
}
