package trakt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nivl/trakt-netflix/internal/errutil"
	"github.com/Nivl/trakt-netflix/internal/pathutil"
	"github.com/Nivl/trakt-netflix/internal/secret"
)

// ErrPendingAuthorization is returned when the authorization is still
// pending, waiting for the user to complete the authorization flow.
var ErrPendingAuthorization = errors.New("pending authorization.")

// TraktAccessToken represents the access token structure returned by
// the Trakt API after a successful authentication or token refresh.
type TraktAccessToken struct {
	AccessToken  secret.Secret `json:"access_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	RefreshToken secret.Secret `json:"refresh_token"`
	Scope        string        `json:"scope"`
	CreatedAt    int64         `json:"created_at"`
}

// Client is the main struct for interacting with the Trakt API.
type Client struct {
	// http is the HTTP client used to make requests to the Trakt API.
	http *http.Client
	// baseURL is the base URL for the Trakt API.
	baseURL string
	// clientID is the client ID of the Trakt APP.
	clientID string
	// clientSecret is the client secret of the Trakt APP.
	clientSecret secret.Secret
	// redirectURI is the redirect URI for the Trakt APP.
	redirectURI string

	auth         TraktAccessToken
	authFilePath string
}

// ClientConfig holds the configuration for the Trakt client.
type ClientConfig struct {
	ClientSecret    secret.Secret `env:"CLIENT_SECRET"`
	ClientID        string        `env:"CLIENT_ID,required"`
	RedirectURI     string        `env:"REDIRECT_URI,required"`
	RelAuthFilePath string        `env:"AUTH_FILE_REL_PATH"`
}

// NewClient creates a new Trakt API client with the provided configuration.
// It reads the authentication tokens from the specified file path.
func NewClient(cfg ClientConfig) (clt *Client, err error) {
	if cfg.RelAuthFilePath == "" {
		cfg.RelAuthFilePath = "trakt_auth.json"
	}

	authFilePath := filepath.Join(pathutil.ConfigDir(), cfg.RelAuthFilePath)

	// TODO(melvin): Use something more secure than ReadFile, to avoid
	// loading a huge file in memory.
	f, err := os.Open(authFilePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("open auth file: %w", err)
	}
	var authTokens TraktAccessToken
	if !os.IsNotExist(err) {
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&authTokens); err != nil {
			return nil, fmt.Errorf("decode auth file: %w", err)
		}
	}

	return &Client{
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:      "https://api.trakt.tv/",
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURI:  cfg.RedirectURI,
		authFilePath: filepath.Join(pathutil.ConfigDir(), cfg.RelAuthFilePath),
		auth:         authTokens,
	}, nil
}

// requestOptions holds options for the request
type requestOptions struct {
	dontRetryOnAuthFailure bool
	noAuth                 bool
}

type requestOptionsFunc func(*requestOptions)

// withNoRetryOnForbidden is a option for the request that indicates
// that the request should not be retried if it receives a 403 Forbidden
// response.
func withNoRetryOnAuthFailure() requestOptionsFunc {
	return func(opts *requestOptions) {
		opts.dontRetryOnAuthFailure = true
	}
}

// withNoAuth is a option for the request that indicates
// that the request should not include authentication headers.
func withNoAuth() requestOptionsFunc {
	return func(opts *requestOptions) {
		opts.noAuth = true
	}
}

// request sends an HTTP request to the Trakt API and returns the response.
// It handles the authentication automatically, refreshing the access token
// if it has expired or is invalid.
func (c *Client) request(ctx context.Context, method string, path string, body json.RawMessage, opts ...requestOptionsFunc) (resp *http.Response, respBody []byte, err error) {
	var options requestOptions
	for _, o := range opts {
		o(&options)
	}

	var bodyBuffer io.Reader = http.NoBody
	if body != nil {
		bodyBuffer = bytes.NewBuffer(body)
	}

	resp, respBody, err = c._request(ctx, method, path, bodyBuffer, options)
	if err != nil {
		return nil, nil, err
	}
	if !options.noAuth && !options.dontRetryOnAuthFailure && resp.StatusCode == http.StatusUnauthorized {
		newTokens, err := c.RefreshToken(ctx, c.auth.RefreshToken.Get())
		if err != nil {
			return nil, nil, fmt.Errorf("refresh token: %w", err)
		}
		c.auth = newTokens.TraktAccessToken
		if err = c.WriteAuthFile(); err != nil {
			return nil, nil, fmt.Errorf("write auth file on disk: %w", err)
		}
		newOpts := append(opts, withNoRetryOnAuthFailure())
		return c.request(ctx, method, path, body, newOpts...)
	}

	return resp, respBody, nil
}

// _request is a low-level HTTP request function that sends a request to the
// Trakt API and returns the response and body.
// It is used internally by the Client methods to handle the actual HTTP
// communication.
func (c *Client) _request(ctx context.Context, method string, path string, body io.Reader, options requestOptions) (resp *http.Response, respBody []byte, err error) {
	if strings.HasSuffix(c.baseURL, "/") && strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("create new HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", c.clientID)
	if !options.noAuth {
		req.Header.Set("Authorization", "Bearer "+c.auth.AccessToken.Get())
	}

	resp, err = c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("send HTTP request: %w", err)
	}
	defer errutil.RunAndSetError(resp.Body.Close, &err, "close response body")

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response body: %w", err)
	}

	return resp, respBody, nil
}

func (c *Client) post(ctx context.Context, path string, body any, opts ...requestOptionsFunc) (resp *http.Response, respBody []byte, err error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal the body: %w", err)
	}

	return c.request(ctx, http.MethodPost, path, jsonBody, opts...)
}

func (c *Client) get(ctx context.Context, path string, opts ...requestOptionsFunc) (resp *http.Response, respBody []byte, err error) {
	return c.request(ctx, http.MethodGet, path, nil, opts...)
}

// GenerateAuthCodeRequest contains the request body for the
// GenerateAuthCode method.
type GenerateAuthCodeRequest struct {
	ClientID string `json:"client_id"`
}

// GenerateAuthCodeResponse contains the response from the
// GenerateAuthCode method.
type GenerateAuthCodeResponse struct {
	DeviceCode      string        `json:"device_code"`
	UserCode        string        `json:"user_code"`
	VerificationURL string        `json:"verification_url"`
	ExpiresIn       int           `json:"expires_in"`
	IntervalInSecs  time.Duration `json:"interval"`
}

// GenerateAuthCode generates an authentication code for the user to
// authorize the application.
func (c *Client) GenerateAuthCode(ctx context.Context) (*GenerateAuthCodeResponse, error) {
	resp, body, err := c.post(ctx, "/oauth/device/code", &GenerateAuthCodeRequest{
		ClientID: c.clientID,
	}, withNoAuth())
	if err != nil {
		return nil, fmt.Errorf("generate auth code: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d. See https://trakt.docs.apiary.io/#introduction/status-codes", resp.StatusCode)
	}

	var authCodeResp GenerateAuthCodeResponse
	if err := json.Unmarshal(body, &authCodeResp); err != nil {
		return nil, err
	}

	return &authCodeResp, nil
}

// GetAccessTokenRequest contains the request body for the GetAccessToken
// method.
type GetAccessTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	DeviceCode   string `json:"code"`
}

// GetAccessTokenResponse contains the response from the GetAccessToken method.
type GetAccessTokenResponse struct {
	TraktAccessToken
}

// GetAccessToken retrieves the access token using the device code
// obtained from GenerateAuthCode.
//
// Returns ErrPendingAuthorization if the authorization is still pending.
// The caller needs to continue polling until the this method returns
// something else.
//
// Once retrieved, the access token is automatically written to the
// auth file on disk.
func (c *Client) GetAccessToken(ctx context.Context, deviceCode string) (*GetAccessTokenResponse, error) {
	// https://trakt.docs.apiary.io/#reference/authentication-devices/get-token/poll-for-the-access_token
	resp, body, err := c.post(ctx, "/oauth/device/token", &GetAccessTokenRequest{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret.Get(),
		DeviceCode:   deviceCode,
	}, withNoAuth())
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, ErrPendingAuthorization
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d. See https://trakt.docs.apiary.io/#introduction/status-codes", resp.StatusCode)
	}

	var accessTokenResp GetAccessTokenResponse
	if err = json.Unmarshal(body, &accessTokenResp); err != nil {
		return nil, err
	}

	c.auth = accessTokenResp.TraktAccessToken
	if err = c.WriteAuthFile(); err != nil {
		return nil, fmt.Errorf("write auth file on disk: %w", err)
	}

	return &accessTokenResp, nil
}

// RefreshTokenRequest contains the request body for the
// RefreshToken method.
type RefreshTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	RedirectURI  string `json:"redirect_uri"`
	GrantType    string `json:"grant_type"`
}

// RefreshTokenResponse contains the response from the RefreshToken method,
// which includes the new access token.
type RefreshTokenResponse struct {
	TraktAccessToken
}

// RefreshToken refreshes the access token using the refresh token.
// Once refreshed, the access token is automatically written to the
// auth file on disk.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	resp, body, err := c.post(ctx, "/oauth/token", &RefreshTokenRequest{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret.Get(),
		RedirectURI:  c.redirectURI,
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}, withNoAuth(), withNoRetryOnAuthFailure())
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d. See https://trakt.docs.apiary.io/#introduction/status-codes", resp.StatusCode)
	}

	var refreshTokenResp RefreshTokenResponse
	if err = json.Unmarshal(body, &refreshTokenResp); err != nil {
		return nil, err
	}

	c.auth = refreshTokenResp.TraktAccessToken
	if err = c.WriteAuthFile(); err != nil {
		return nil, fmt.Errorf("write auth file on disk: %w", err)
	}

	return &refreshTokenResp, nil
}

// SearchResponse contains the response from the Search method.
type SearchResponse struct {
	Results []struct {
		Type    SearchTypes `json:"type"`
		Movie   Media       `json:"movie,omitempty"`
		Episode Episode     `json:"episode,omitempty"`
		Show    Media       `json:"show,omitempty"`
	} `json:"results"`
}

// Search searches for a media item on Trakt using the provided query.
func (c *Client) Search(ctx context.Context, typ SearchTypes, query string) (*SearchResponse, error) {
	url := fmt.Sprintf("/search/%s?query=%s", typ, query)

	resp, body, err := c.get(ctx, url, withNoAuth())
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d. See https://trakt.docs.apiary.io/#introduction/status-codes", resp.StatusCode)
	}

	var searchResponse SearchResponse
	if err = json.Unmarshal(body, &searchResponse.Results); err != nil {
		return nil, err
	}

	return &searchResponse, nil
}

type MarkAsWatchedRequest struct {
	Movies   []MarkAsWatched `json:"movies"`
	Episodes []MarkAsWatched `json:"episodes"`
}

type MarkAsWatchedResponse struct {
	Added struct {
		Movies   int `json:"movies,omitempty"`
		Episodes int `json:"episodes,omitempty"`
	} `json:"added"`
	NotFound struct {
		Movies []struct {
			IDs IDs `json:"ids"`
		} `json:"movies,omitempty"`
		Episodes []struct {
			IDs IDs `json:"ids"`
		} `json:"episodes,omitempty"`
	} `json:"not_found"`
}

// MarkAsWatched marks a media item as watched on Trakt.
func (c *Client) MarkAsWatched(ctx context.Context, req *MarkAsWatchedRequest) (*MarkAsWatchedResponse, error) {
	resp, body, err := c.post(ctx, "/sync/history", req)
	if err != nil {
		return nil, fmt.Errorf("mark as watched: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("http %d. See https://trakt.docs.apiary.io/#introduction/status-codes", resp.StatusCode)
	}

	var response MarkAsWatchedResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// unsecuredTraktAccessToken is a struct that contains the access token
// and refresh token in plain text, so that it can be written to the
// auth file on disk.
type unsecuredTraktAccessToken struct {
	TraktAccessToken
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// WriteAuthFile writes the current authentication data to the
// auth file on disk.
func (c *Client) WriteAuthFile() error {
	auth := unsecuredTraktAccessToken{
		TraktAccessToken: c.auth,
		AccessToken:      c.auth.AccessToken.Get(),
		RefreshToken:     c.auth.RefreshToken.Get(),
	}
	data, err := json.Marshal(auth)
	if err != nil {
		return fmt.Errorf("marshal auth data: %w", err)
	}
	return os.WriteFile(c.authFilePath, data, 0o600)
}
