package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/runtipi/cli/internal/assets"
	"github.com/runtipi/cli/internal/config"
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

func CopySystemFiles() error {
	dockerComposeContent := assets.DockerCompose

	err := os.WriteFile(filepath.Join(config.RootFolder, "docker-compose.yml"), dockerComposeContent, 0664)
	if err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	versionContent := assets.Version

	err = os.WriteFile(filepath.Join(config.RootFolder, "VERSION"), versionContent, 0664)
	if err != nil {
		return fmt.Errorf("failed to write VERSION file: %w", err)
	}

	// Create all required directories
	directories := []string{
		"apps",
		"data",
		"app-data",
		"state",
		"repos",
		"media",
		"traefik",
		"user-config",
		"logs",
		"backups",
	}

	for _, dir := range directories {
		dirPath := filepath.Join(config.RootFolder, dir)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	traefikSharedDir := filepath.Join(config.RootFolder, "traefik", "shared")
	err = os.MkdirAll(traefikSharedDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create traefik/shared directory: %w", err)
	}

	return nil
}

func EnsureFilePermissions() error {
	permissionItems := []struct {
		perms string
		paths []string
	}{
		{"777", []string{"state", "data", "apps", "logs", "traefik", "repos", "user-config", "state"}},
		{"666", []string{"state/settings.json"}},
		{"664", []string{".env", "docker-compose.yml", "VERSION"}},
		{"600", []string{"traefik/shared/acme.json", "state/seed"}},
	}

	for _, item := range permissionItems {
		for _, path := range item.paths {
			fullPath := filepath.Join(config.RootFolder, path)

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				continue
			}

			cmd := exec.Command("chmod", "-Rf", item.perms, fullPath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s has incorrect permissions. Please run the CLI as root to fix this. Error: %s, Output: %s",
					path, err, string(output))
			}
		}
	}

	return nil
}
