package keychain

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	serviceName = "confluence-cli"
)

// StorePAT stores a Personal Access Token in the macOS Keychain.
func StorePAT(account, token string) error {
	// First try to delete any existing entry (ignore errors if it doesn't exist)
	_ = exec.Command("security", "delete-generic-password",
		"-s", serviceName,
		"-a", account,
	).Run()

	cmd := exec.Command("security", "add-generic-password",
		"-s", serviceName,
		"-a", account,
		"-w", token,
		"-U",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to store PAT in keychain: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// GetPAT retrieves a Personal Access Token from the macOS Keychain.
func GetPAT(account string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", serviceName,
		"-a", account,
		"-w",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to read PAT from keychain (have you run 'confluence auth store'?): %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// DeletePAT removes a Personal Access Token from the macOS Keychain.
func DeletePAT(account string) error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", serviceName,
		"-a", account,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete PAT from keychain: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}
