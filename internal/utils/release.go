package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/runtipi/cli/internal/config"
)

var NonPreReleaseTags = []string{"nightly", "e2e"}

type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Prerelease  bool   `json:"prerelease"`
	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name        string `json:"name"`
		ContentType string `json:"content_type"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetReleases(url string) ([]GitHubRelease, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Runtipi-CLI")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases. Status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// The latest release is not returned in an array so we will try to decode it as a single release
	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		var release GitHubRelease
		if err := json.Unmarshal(body, &release); err != nil {
			return nil, fmt.Errorf("failed to parse releases: %v", err)
		}
		releases = append(releases, release)
	}

	return releases, nil
}

func IsMajorBump(currentVersion, newVersion string) bool {
	if newVersion == "nightly" {
		return false
	}

	currentVersion = strings.TrimPrefix(currentVersion, "v")
	newVersion = strings.TrimPrefix(newVersion, "v")

	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false
	}

	new, err := semver.NewVersion(newVersion)
	if err != nil {
		return false
	}

	return current.Major() < new.Major()
}

func DownloadReleaseAndSelfReplace(version string) error {
	release, err := FindReleaseByVersion(version)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	arch := runtime.GOARCH
	switch arch {
	case "arm64":
		arch = "aarch64"
	case "amd64":
		arch = "x86_64"
	}

	var assetURL string
	var assetName string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, arch) && strings.Contains(asset.Name, runtime.GOOS) {
			assetURL = asset.DownloadURL
			assetName = asset.Name
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no asset found for %s %s on release %s", arch, runtime.GOOS, release.TagName)
	}

	tempDir, err := os.MkdirTemp(config.RootFolder, "self_update")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Runtipi-CLI")
	req.Header.Set("Accept", "application/octet-stream")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download release. Status code: %d", resp.StatusCode)
	}

	tempFilePath := filepath.Join(tempDir, assetName)
	file, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	_, err = io.Copy(file, resp.Body)
	file.Close()
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	err = ExtractTarGz(tempFilePath, config.RootFolder)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %v", err)
	}

	binName := strings.Split(assetName, ".")[0]
	newExecutablePath := filepath.Join(config.RootFolder, binName)

	err = os.Chmod(newExecutablePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to make binary executable: %v", err)
	}

	err = os.Rename(newExecutablePath, filepath.Join(config.RootFolder, "runtipi-cli"))
	if err != nil {
		return fmt.Errorf("could not rename old binary: %w", err)
	}

	return nil
}

func FindReleaseByVersion(version string) (GitHubRelease, error) {
	// https://api.github.com/repos/runtipi/runtipi/releases/tags/v4.0.0
	url := fmt.Sprintf("https://api.github.com/repos/runtipi/runtipi/releases/tags/%s", version)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Runtipi-CLI")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var release GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return GitHubRelease{}, fmt.Errorf("failed to parse release: %v", err)
		}
		return release, nil
	}

	return GitHubRelease{}, fmt.Errorf("release not found. Did you forget the v prefix? (e.g. v4.0.0 instead of 4.0.0)")
}

func FilterNonPreReleases(releases []GitHubRelease) []GitHubRelease {
	var filtered []GitHubRelease
	for _, release := range releases {
		if slices.Contains(NonPreReleaseTags, release.TagName) || !release.Prerelease {
			continue
		}
		filtered = append(filtered, release)
	}
	return filtered
}
