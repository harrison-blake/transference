package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/harrison-blake/envreader"
	api "github.com/harrison-blake/transference/api"
	auth "github.com/harrison-blake/transference/auth"
)

func main() {
	if err := envreader.Load("./.env"); err != nil {
		log.Fatalf("FATAL: could not load .env file: %v", err)
	}
	//**************************************************
	//                  SPOTIFY AUTH
	//**************************************************
	spotifyAuthenticator, err := auth.NewSpotifyAuthenticator()
	if err != nil {
		log.Fatalf("failed to create spotify authenticator: %v", err)
	}

	if err := spotifyAuthenticator.PerformAuthFlow(); err != nil {
		log.Fatalf("Spotify authentication failed: %v", err)
	}

	fmt.Print("Successfully authenticated with Spotify\n")
	//**************************************************
	//                YOUTUBE AUTH
	//**************************************************
	youtubeAuthenticator, err := auth.NewYoutubeAuthenticator()
	if err != nil {
		log.Fatalf("failed to create youtube authenticator: %v", err)
	}

	if err := youtubeAuthenticator.PerformAuthFlow(); err != nil {
		log.Fatalf("Youtube authentication failed: %v", err)
	}

	fmt.Print("Successfully authenticated with Youtube\n")
	//**************************************************
	//          SPOTIFY GET USERS PLAYLISTS
	//**************************************************
	spotifyClient := api.NewClient(spotifyAuthenticator.GetToken())
	userPlaylists, err := spotifyClient.GetUserPlaylists()
	if err != nil {
		log.Fatalf("failed to get user playlists: %v", err)
	}

	selectedPlaylist := ChoosePlaylist(userPlaylists)

	if selectedPlaylist == nil {
		log.Fatalf("Playlist not found")
	}

	fmt.Printf("You selected: %s (ID: %s)\n", selectedPlaylist.Name, selectedPlaylist.ID)
	//**************************************************
	//          SPOTIFY GET PLAYLIST TRACKS
	//**************************************************
	playlistTracks, err := spotifyClient.GetPlaylistTracks(selectedPlaylist.ID)
	if err != nil {
		log.Fatalf("failed to get playlist tracks: %v", err)
	}

	fmt.Println("\nTracks in selected playlist:")
	for _, item := range playlistTracks.Items {
		fmt.Printf("\n- %s\n", item.Track.Name)
		fmt.Println("  Artists:")
		for _, artist := range item.Track.Artists {
			fmt.Printf("    - %s\n", artist.Name)
		}
	}
	//**************************************************
	//          CREATE YOUTUBE PLAYLIST
	//**************************************************
	youtubeClient := api.NewClient(youtubeAuthenticator.GetToken())
	youtubePlaylist, err := youtubeClient.CreatePlaylist()
	if err != nil {
		log.Fatalf("failed to create playlist: %v\n", err)
	}

	fmt.Print("Successfully created playlist\n")
	//**************************************************
	//          ADD TRACKS TO YOUTUBE PLAYLIST
	//**************************************************
	for _, track := range playlistTracks.Items {
		songName := track.Track.Name
		artistName := track.Track.Artists[0].Name
		searchResponse, _ := youtubeClient.YoutubeSearch(songName, artistName)
		plItemResource, err := youtubeClient.AddSong(youtubePlaylist.ID, searchResponse.Items[0].ID["videoId"])
		if err != nil {
			log.Fatalf("failed to add to playlist: %v\n", err)
		}
		fmt.Printf("resource kind: %v\n", plItemResource.Kind)
	}

	fmt.Print("Playlist successfully transfered\n")
}

func ChoosePlaylist(playlists *api.UserPlaylists) *api.Playlist {
	fmt.Println("\nAvailable Playlists:")
	for _, p := range playlists.Playlists {
		fmt.Printf("- %s\n", p.Name)
	}

	fmt.Print("\nEnter the name of the playlist you want to copy: ")
	reader := bufio.NewReader(os.Stdin)
	selectedPlaylistName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("failed to read user input: %v", err)
	}
	selectedPlaylistName = strings.TrimSpace(selectedPlaylistName)

	for _, p := range playlists.Playlists {
		if strings.EqualFold(p.Name, selectedPlaylistName) {
			return &p
		}
	}

	return nil
}