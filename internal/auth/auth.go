package auth

import (
	"fmt"
	"os"

	"github.com/bentsolheim/confluence-cli/internal/keychain"
)

// ResolveToken attempts to find a Confluence PAT using these sources (in order):
//  1. macOS Keychain (keyed by confluenceURL)
//  2. CONFLUENCE_PAT environment variable
func ResolveToken(confluenceURL string) (string, error) {
	// Try keychain first
	token, err := keychain.GetPAT(confluenceURL)
	if err == nil && token != "" {
		return token, nil
	}

	// Fall back to environment variable
	if envToken := os.Getenv("CONFLUENCE_PAT"); envToken != "" {
		return envToken, nil
	}

	return "", fmt.Errorf("no authentication token found: try 'confluence auth store' or set CONFLUENCE_PAT environment variable")
}
