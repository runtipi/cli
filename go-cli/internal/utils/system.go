package utils

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// EnsureDocker checks if Docker is available and the user has proper permissions
func EnsureDocker() error {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	outputBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get docker version: %w", err)
	}
	version := strings.TrimSpace(string(outputBytes))
	major := strings.Split(version, ".")[0]

	majorVersion, err := strconv.Atoi(major)
	if err != nil {
		return fmt.Errorf("failed to parse docker version: %w", err)
	}

	if majorVersion < MinimumDockerVersion {
		return fmt.Errorf("docker version %s is not supported, please update to at least version %d", version, MinimumDockerVersion)
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

