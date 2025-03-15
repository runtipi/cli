package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateSeed creates a new seed file if it doesn't exist
func GenerateSeed(rootDir string) error {
	seedPath := filepath.Join(rootDir, "state", "seed")

	// Check if seed file already exists
	if _, err := os.Stat(seedPath); err == nil {
		return nil
	}

	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	seed := hex.EncodeToString(randomBytes)

	// Create state directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(rootDir, "state"), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Write seed file
	if err := os.WriteFile(seedPath, []byte(seed), 0644); err != nil {
		return fmt.Errorf("failed to write seed file: %w", err)
	}

	return nil
}

