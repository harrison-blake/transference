package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	Token  string
	Client *http.Client
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserPlaylists struct {
	Playlists []Playlist `json:"items"`
}

type Track struct {
	Name   string `json:"name"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
}

type PlaylistTrackItem struct {
	Track Track `json:"track"`
}

type PlaylistTracksResponse struct {
	Items []PlaylistTrackItem `json:"items"`
}

func NewClient(token string) *Client {
	return &Client{
		Token:  token,
		Client: &http.Client{},
	}
}

func (s *Client) GetUserPlaylists() (*UserPlaylists, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/playlists", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlists request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer " + s.Token)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send users playlist request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users playlist request request failed with status: %s", resp.Status)
	}

	var userPlaylists UserPlaylists
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(body, &userPlaylists)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	return &userPlaylists, nil
}

func (s *Client) GetPlaylistTracks(playlistID string) (*PlaylistTracksResponse, error) {
	baseURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	params := url.Values{}
	params.Add("market", "ES")
	params.Add("fields", "items(track(name, artists(name)))")
	params.Add("limit", "100")
	parsedURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist tracks request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer " + s.Token)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send playlist tracks request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("playlist tracks request failed with status: %s, URL: %s, Body: %s", resp.Status, parsedURL.String(), string(bodyBytes))
	}

	var playlistTracks PlaylistTracksResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &playlistTracks); err != nil {
		return nil, fmt.Errorf("failed to decode playlist tracks response: %w", err)
	}

	return &playlistTracks, nil
}