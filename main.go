// 🤖 AI-generated
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	var clientID, clientSecret, scopes string
	flag.StringVar(&clientID, "client-id", "", "Spotify application client ID")
	flag.StringVar(&clientSecret, "client-secret", "", "Spotify application client secret")
	flag.StringVar(&scopes, "scopes", "", "Comma-separated list of OAuth scopes")
	flag.Parse()

	if clientID == "" || clientSecret == "" || scopes == "" {
		fmt.Fprintln(os.Stderr, "Usage: spotify-auth --client-id=ID --client-secret=SECRET --scopes=SCOPES")
		fmt.Fprintln(os.Stderr, "All three flags are required.")
		os.Exit(1)
	}

	cfg := Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       strings.TrimSpace(strings.ReplaceAll(scopes, ",", " ")),
		RedirectURI:  callbackURI,
	}

	state, err := generateState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	resultCh, srv := startServer(state)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx) //nolint
	}()

	authURL := buildAuthURL(cfg, state)

	fmt.Fprintln(os.Stderr, "Opening browser for Spotify authorization...")
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please visit: %s\n", authURL)
	}
	fmt.Fprintf(os.Stderr, "Waiting for callback on %s\n", callbackURI)

	result := <-resultCh

	if result.err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", result.err)
		os.Exit(1)
	}

	tok, err := exchangeCode(cfg, result.code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAccess token (use this for immediate API calls, expires in 1h):\n%s\n\nRefresh token (paste this into your app config):\n%s\n", tok.AccessToken, tok.RefreshToken)
}
