package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type VideoDetails struct {
	Kind string            `json:"kind"`
	Etag string            `json:"etag"`
	ID   map[string]string `json:"id"`
}

type SearchResponse struct {
	Kind          string         `json:"kind"`
	Etag          string         `json:"etag"`
	NextPageToken string         `json:"nextPageToken"`
	RegionCode    string         `json:"regionCode"`
	PageInfo      map[string]int `json:"pageInfo"`
	Items         []VideoDetails `json:"items"`
}

type PlaylistDetails struct {
	PublishedAt string `json:"publishdAt"`
	ChannelId   string `json:"channelId"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type PlaylistCreationResponse struct {
	Kind    string          `json:"kind"`
	Etag    string          `json:"items"`
	ID      string          `json:"id"`
	Details PlaylistDetails `json:"snippet"`
}

type PlaylistItemResource struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		PublishedAt  string                 `json:"publishedAt"`
		ChannelID    string                 `json:"channelId"`
		Title        string                 `json:"title"`
		Description  string                 `json:"description"`
		Thumbnails   map[string]interface{} `json:"thumbnails"`
		ChannelTitle string                 `json:"channelTitle"`
		PlaylistID   string                 `json:"playlistId"`
		Position     int                    `json:"position"`
		ResourceID   struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"resourceId"`
	} `json:"snippet"`
}

func (c *Client) YoutubeSearch(songName string, artistName string) (*SearchResponse, error) {
	parsedURL, _ := url.Parse("https://www.googleapis.com/youtube/v3/search")
	params := url.Values{}
	params.Add("part", "snippet")
	params.Add("q", songName+" "+artistName)
	params.Add("maxResults", "1")
	parsedURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Youtube search request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Youtube search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube search request failed with status: %s", resp.Status)
	}

	var searchResponse SearchResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode playlist tracks response: %w", err)
	}

	return &searchResponse, nil
}

func (c *Client) CreatePlaylist() (*PlaylistCreationResponse, error) {
	var postBody = `{
	  "snippet": {
	    "title": "My New Playlist",
	    "description": "A playlist created using the YouTube API",
	    "tags": ["music", "favorites"],
	    "defaultLanguage": "en"
	  },
	  "status": {
	    "privacyStatus": "public"
	  }
	}`

	parsedURL, _ := url.Parse("https://www.googleapis.com/youtube/v3/playlists")
	params := url.Values{}
	params.Add("part", "snippet,status")
	parsedURL.RawQuery = params.Encode()

	req, err := http.NewRequest("POST", parsedURL.String(), strings.NewReader(postBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send playlist request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("playlist request failed with status: %s, URL: %s", resp.Status, parsedURL.String())
	}

	var plResponse PlaylistCreationResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(body, &plResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &plResponse, nil
}

func (c *Client) AddSong(playlistID string, videoID string) (*PlaylistItemResource, error) {
	var postBody = `{
    "snippet": {
        "playlistId": "` + playlistID + `",
        "resourceId": {
            "kind": "youtube#video",
            "videoId": "` + videoID + `"
        }
    }
    }`

	parsedURL, _ := url.Parse("https://www.googleapis.com/youtube/v3/playlistItems")
	params := url.Values{}
	params.Add("part", "snippet")
	parsedURL.RawQuery = params.Encode()

	req, err := http.NewRequest("POST", parsedURL.String(), strings.NewReader(postBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send playlist request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("playlist request failed with status: %s, URL: %s, Body: %s", resp.Status, parsedURL.String(), string(body))
	}

	var playlistItemResource PlaylistItemResource
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(body, &playlistItemResource)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &playlistItemResource, nil
}