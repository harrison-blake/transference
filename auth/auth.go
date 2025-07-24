package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const tokenFilePath = "auth.json"

type Config struct {
	ClientID      string
	ClientSecret  string
	AuthEndpoint  string
	TokenEndpoint string
	RedirectURI   string
	EncodedClient string
	Callback      string
	AuthURL       *url.URL
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type Authenticator struct {
	Conf   *Config
	Token  *TokenResponse
	Client *http.Client
}

func (a *Authenticator) GetToken() string {
	return a.Token.AccessToken
}

func DefaultSpotifyConifg() (*Config, error) {
	conf := &Config{
		ClientID:      os.Getenv("SPOTIFY_ID"),
		ClientSecret:  os.Getenv("SPOTIFY_SECRET"),
		AuthEndpoint:  os.Getenv("SPOTIFY_AUTH_ENDPOINT"),
		TokenEndpoint: os.Getenv("SPOTIFY_TOKEN_ENDPOINT"),
		RedirectURI:   os.Getenv("SPOTIFY_REDIRECT_URI"),
	}

	authURL, err := url.Parse(conf.AuthEndpoint)
	if err != nil {
		return nil, err
	}
	conf.AuthURL = authURL

	redURL, err := url.Parse(conf.RedirectURI)
	if err != nil {
		return nil, err
	}
	conf.Callback = redURL.Path

	conf.setDefaultSpotifyURLValues()
	conf.encodeClient()

	return conf, nil
}

func DefaultYoutubeConifg() (*Config, error) {
	conf := &Config{
		ClientID:      os.Getenv("YOUTUBE_ID"),
		ClientSecret:  os.Getenv("YOUTUBE_SECRET"),
		AuthEndpoint:  os.Getenv("YOUTUBE_AUTH_ENDPOINT"),
		TokenEndpoint: os.Getenv("YOUTUBE_TOKEN_ENDPOINT"),
		RedirectURI:   os.Getenv("YOUTUBE_REDIRECT_URI"),
		Callback:      os.Getenv("YOUTUBE_CALLBACK_PATH"),
	}

	authURL, err := url.Parse(conf.AuthEndpoint)
	if err != nil {
		return nil, err
	}
	conf.AuthURL = authURL

	redURL, err := url.Parse(conf.RedirectURI)
	if err != nil {
		return nil, err
	}
	conf.Callback = redURL.Path

	conf.setDefaultYoutubeURLValues()
	conf.encodeClient()

	return conf, nil
}

func NewSpotifyAuthenticator() (*Authenticator, error) {
	conf, err := DefaultSpotifyConifg()
	if err != nil {
		return nil, err
	}

	return &Authenticator{Conf: conf, Client: &http.Client{}}, nil
}

func NewYoutubeAuthenticator() (*Authenticator, error) {
	conf, err := DefaultYoutubeConifg()
	if err != nil {
		return nil, err
	}

	return &Authenticator{Conf: conf, Client: &http.Client{}}, nil
}

func (a *Authenticator) PerformAuthFlow() error {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	server := &http.Server{Addr: ":8080"}
	var serverErr error

	http.HandleFunc(a.Conf.Callback, func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		code := r.URL.Query().Get("code")
		if code == "" {
			serverErr = fmt.Errorf("authorization code not found in callback")
			return
		}

		if err := a.exchangeCodeForToken(code); err != nil {
			serverErr = fmt.Errorf("failed to exchange code for token: %w", err)
			return
		}

		fmt.Fprintf(w, "Authentication successful! You can close this window.")

		go func() {
			if err := server.Shutdown(context.Background()); err != nil {
				log.Printf("Failed to shut down server: %v", err)
			}
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	cmd := exec.Command("open", a.Conf.GetAuthURL())
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	wg.Wait()

	// Save the token after successful acquisition
	buffer, err := json.Marshal(a.Token)
	if err != nil {
		return fmt.Errorf("failed to marshal token for saving: %w", err)
	}

	if err := os.WriteFile(tokenFilePath, buffer, 0600); err != nil {
		return fmt.Errorf("failed to save token to file: %w", err)
	}

	return serverErr
}

func (a *Authenticator) exchangeCodeForToken(code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", a.Conf.RedirectURI)

	req, err := http.NewRequest("POST", a.Conf.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic " + a.Conf.EncodedClient)

	resp, err := a.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status: %s", resp.Status)
	}

	var tokenResponse TokenResponse
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	a.Token = &tokenResponse
	return nil
}

func (c *Config) setDefaultSpotifyURLValues() {
	values := url.Values{}
	values.Add("response_type", "code")
	values.Add("client_id", c.ClientID)
	values.Add("scope", "playlist-read-private playlist-read-collaborative")
	values.Add("redirect_uri", c.RedirectURI)
	c.AuthURL.RawQuery = values.Encode()
}

func (c *Config) setDefaultYoutubeURLValues() {
	values := url.Values{}
	values.Add("response_type", "code")
	values.Add("client_id", c.ClientID)
	values.Add("scope", "https://www.googleapis.com/auth/youtube")
	values.Add("redirect_uri", c.RedirectURI)
	c.AuthURL.RawQuery = values.Encode()
}

func (c *Config) encodeClient() {
	if c.ClientID != "" && c.ClientSecret != "" {
		c.EncodedClient = base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))
	}
}

func (c *Config) GetAuthURL() string {
	return c.AuthURL.String()
}
