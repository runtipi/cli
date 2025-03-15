package utils

import (
	"fmt"
	"os/exec"
)

// EnsureDocker checks if Docker is available and the user has proper permissions
func EnsureDocker() error {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not running or user doesn't have proper permissions: %w", err)
	}
	return nil
}

// CopySystemFiles copies necessary system files
func CopySystemFiles() error {
	// TODO: Implement system file copying logic
	return nil
}

// EnsureFilePermissions ensures proper file permissions
func EnsureFilePermissions() error {
	// TODO: Implement file permission logic
	return nil
} 