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
	CreatedAt    int           `json:"created_at"`
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
	dontRetryOnForbidden bool
}

type requestOptionsFunc func(*requestOptions)

// withNoRetryOnForbidden is a option for the request that indicates
// that the request should not be retried if it receives a 403 Forbidden
// response.
func withNoRetryOnForbidden() requestOptionsFunc {
	return func(opts *requestOptions) {
		opts.dontRetryOnForbidden = true
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

	resp, respBody, err = c._request(ctx, method, path, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, err
	}
	if !options.dontRetryOnForbidden && resp.StatusCode == http.StatusForbidden {
		newTokens, err := c.RefreshToken(ctx, c.auth.RefreshToken.Get())
		if err != nil {
			return nil, nil, fmt.Errorf("refresh token: %w", err)
		}
		c.auth = newTokens.TraktAccessToken
		if err = c.WriteAuthFile(); err != nil {
			return nil, nil, fmt.Errorf("write auth file on disk: %w", err)
		}
		newOpts := append(opts, withNoRetryOnForbidden())
		return c.request(ctx, method, path, body, newOpts...)
	}

	return resp, respBody, nil
}

// _request is a low-level HTTP request function that sends a request to the
// Trakt API and returns the response and body.
// It is used internally by the Client methods to handle the actual HTTP
// communication.
func (c *Client) _request(ctx context.Context, method string, path string, body io.Reader) (resp *http.Response, respBody []byte, err error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("create new HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("send HTTP request: %w", err)
	}
	defer errutil.RunAndSetError(resp.Body.Close, &err, "close response body")

	if resp.StatusCode == http.StatusForbidden {
		if err = resp.Body.Close(); err != nil {
			return nil, nil, fmt.Errorf("close response body: %w", err)
		}

		return nil, nil, fmt.Errorf("unauthorized. Check your client ID and secret, or refresh your access token if it has expired")
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response body: %w", err)
	}

	return resp, respBody, nil
}

func (c *Client) post(ctx context.Context, path string, body any) (resp *http.Response, respBody []byte, err error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal the body: %w", err)
	}

	return c.request(ctx, http.MethodPost, path, jsonBody)
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
	})
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
	})
	if err != nil {
		return nil, fmt.Errorf("generate auth code: %w", err)
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
	})
	if err != nil {
		return nil, fmt.Errorf("generate auth code: %w", err)
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
	return os.WriteFile(c.authFilePath, data, 0o644)
}
