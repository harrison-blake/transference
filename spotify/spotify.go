package spotify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/harrison-blake/transference/auth"
)

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserPlaylists struct {
	Playlists []Playlist `json:"items"`
}

func GetUserPlaylists(token *auth.TokenResponse) (*UserPlaylists, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/playlists", nil)
	if err != nil {
		log.Fatalf("failed to create playlists request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to send users playlist request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("users playlist request request failed with status: %s", resp.Status)
	}

	var userPlaylists UserPlaylists
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &userPlaylists)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	return &userPlaylists, nil
}
