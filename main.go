package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/harrison-blake/envreader"
	"github.com/harrison-blake/transference/auth"
	"github.com/harrison-blake/transference/spotify"
)

func main() {
	if err := envreader.Load("./.env"); err != nil {
		log.Fatalf("FATAL: could not load .env file: %v", err)
	}
	//**************************************************
	//                  SPOTIFY AUTH
	//**************************************************

	authenticator, err := auth.NewAuthenticator()
	if err != nil {
		log.Fatalf("failed to create authenticator: %v", err)
	}

	if err := authenticator.PerformAuthFlow(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Print("Successfully authenticated\n")

	//**************************************************
	//          SPOTIFY GET USERS PLAYLISTS
	//**************************************************
	userPlaylists, err := spotify.GetUserPlaylists(authenticator.Token)
	if err != nil {
		log.Fatalf("failed to get user playlists: %v", err)
	}

	selectedPlaylist := ChoosePlaylist(userPlaylists)

	if selectedPlaylist == nil {
		log.Fatalf("Playlist not found")
	}

	fmt.Printf("You selected: %s (ID: %s)", selectedPlaylist.Name, selectedPlaylist.ID)

	//**************************************************
	//          SPOTIFY GET PLAYLIST TRACK LIST
	//**************************************************
}

func ChoosePlaylist(playlists *spotify.UserPlaylists) *spotify.Playlist {
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
