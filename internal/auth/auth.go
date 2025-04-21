// Package auth provides authentication functionality for YouTube API access.
// It handles OAuth2 authentication flow, token management, and client setup.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

// ClientConfig represents OAuth2 client configuration structure as provided by Google.
// It contains credentials and endpoints for OAuth2 authentication.
type ClientConfig struct {
	Installed struct {
		ClientID                string `json:"client_id"`                   // OAuth client ID
		ProjectID               string `json:"project_id"`                  // Google Cloud project ID
		AuthURI                 string `json:"auth_uri"`                    // Authorization endpoint
		TokenURI                string `json:"token_uri"`                   // Token endpoint
		AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"` // Certificate URL for the auth provider
		ClientSecret            string `json:"client_secret"`               // OAuth client secret
	} `json:"installed"`
}

// getTokenPath returns the file path where OAuth tokens are stored.
func getTokenPath() string {
	return "youtube-token.json"
}

// Login initiates the OAuth2 authentication flow for YouTube API access.
// It prompts the user to authorize access in a browser and captures the authorization code.
func Login() error {
	config, err := loadClientConfig()
	if err != nil {
		return err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.Installed.ClientID,
		ClientSecret: config.Installed.ClientSecret,
		RedirectURL:  "http://localhost",
		Scopes: []string{
			youtube.YoutubeUploadScope,
			youtube.YoutubeReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}

	authURL := oauthConfig.AuthCodeURL("state")
	fmt.Printf("\nAccess this URL in your browser:\n\n%v\n\n", authURL)
	fmt.Print("Paste the authorization code that appears on the screen: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return fmt.Errorf("unable to read authorization code: %v", err)
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("unable to exchange code for token: %v", err)
	}

	return saveToken(token)
}

// loadClientConfig reads and parses the OAuth client configuration file.
// Returns the parsed client configuration or an error if the file cannot be read or parsed.
func loadClientConfig() (*ClientConfig, error) {
	config := &ClientConfig{}
	configFile := "credentials.json"

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %v", err)
	}

	return config, nil
}

// saveToken persists an OAuth token to the filesystem for future use.
// The token is stored in the file specified by getTokenPath().
func saveToken(token *oauth2.Token) error {
	tokenPath := getTokenPath()
	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to create token file: %v", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// GetClient retrieves the stored OAuth token.
// Returns an error if the token doesn't exist or can't be parsed.
func GetClient() (*oauth2.Token, error) {
	tokenPath := getTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("token not found. Run 'godeogoker login' first: %v", err)
	}

	token := &oauth2.Token{}
	if err := json.Unmarshal(data, token); err != nil {
		return nil, fmt.Errorf("error reading token: %v", err)
	}

	return token, nil
}
