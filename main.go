package main

import (
	"fmt"
	"log"
	// "os"
	// "encoding/json"

	"github.com/harrison-blake/envreader"
	"github.com/harrison-blake/transference/auth"
)

func main() {
	if err := envreader.Load("./.env"); err != nil {
		log.Fatalf("FATAL: could not load .env file: %v", err)
	}

	authenticator, err := auth.NewAuthenticator()
	if err != nil {
		log.Fatalf("failed to create authenticator: %v", err)
	}

	if err := authenticator.PerformAuthFlow(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Print("Successfully authenticated")
}
