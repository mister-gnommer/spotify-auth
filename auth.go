// 🤖 AI-generated
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const tokenEndpoint = "https://accounts.spotify.com/api/token"

// Config holds validated CLI inputs.
type Config struct {
	ClientID     string
	ClientSecret string
	Scopes       string  // space-separated
	RedirectURI  string  // always "http://127.0.0.1:8888/callback"
}

// callbackResult is sent from the HTTP handler to main via a channel.
type callbackResult struct {
	code string
	err  error
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// generateState returns a 128-bit random URL-safe string for CSRF protection.
func generateState() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// buildAuthURL constructs the Spotify authorization URL.
func buildAuthURL(cfg Config, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURI)
	params.Set("scope", cfg.Scopes)
	params.Set("state", state)
	return "https://accounts.spotify.com/authorize?" + params.Encode()
}

// exchangeCode trades an authorization code for tokens at Spotify's token endpoint.
func exchangeCode(cfg Config, code string) (tokenResponse, error) {
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", cfg.RedirectURI)

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		tokenEndpoint,
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("failed to build token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Basic auth uses StdEncoding per the HTTP spec, not URL encoding.
	creds := base64.StdEncoding.EncodeToString([]byte(cfg.ClientID + ":" + cfg.ClientSecret))
	req.Header.Set("Authorization", "Basic "+creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return tokenResponse{}, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, raw)
	}

	var tok tokenResponse
	if err := json.Unmarshal(raw, &tok); err != nil {
		return tokenResponse{}, fmt.Errorf("failed to parse token response: %w", err)
	}
	if tok.Error != "" {
		return tokenResponse{}, fmt.Errorf("spotify error: %s", tok.ErrorDesc)
	}
	if tok.RefreshToken == "" {
		return tokenResponse{}, fmt.Errorf("spotify did not return a refresh token")
	}

	return tok, nil
}
