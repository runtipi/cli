// Package commands
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

type AppResponseBody struct {
	RequestID string `json:"requestId"`
}

type InstalledAppsResponse struct {
	// A pointer lets us distinguish an empty installed-app list from a
	// missing or null field in the API response.
	Installed *[]InstalledApp `json:"installed"`
}

type InstalledApp struct {
	Info     AppInfo     `json:"info"`
	App      AppDetails  `json:"app"`
	Metadata AppMetadata `json:"metadata"`
}

type AppInfo struct {
	URN     *string `json:"urn"`
	Version *string `json:"version"`
}

type AppDetails struct {
	Version *int `json:"version"`
}

type AppMetadata struct {
	LatestVersion       *int    `json:"latestVersion"`
	LatestDockerVersion *string `json:"latestDockerVersion"`
}

// AppUpdate describes an app with a newer Tipi version available.
type AppUpdate struct {
	URN            string
	CurrentVersion string
	LatestVersion  string
	TipiVersion    int
	LatestTipi     int
}

// FindAvailableUpdates validates the installed-app response and returns apps
// whose Tipi version is behind the latest available version.
func FindAvailableUpdates(response InstalledAppsResponse) ([]AppUpdate, error) {
	if response.Installed == nil {
		return nil, fmt.Errorf("unexpected response format: 'installed' field is missing or null")
	}

	updates := make([]AppUpdate, 0)
	for index, app := range *response.Installed {
		if app.Info.URN == nil || strings.TrimSpace(*app.Info.URN) == "" {
			return nil, fmt.Errorf("unexpected response format: installed[%d].info.urn is missing or empty", index)
		}
		if app.Info.Version == nil || strings.TrimSpace(*app.Info.Version) == "" {
			return nil, fmt.Errorf("unexpected response format: installed[%d].info.version is missing or empty", index)
		}
		if app.App.Version == nil {
			return nil, fmt.Errorf("unexpected response format: installed[%d].app.version is missing", index)
		}
		if app.Metadata.LatestVersion == nil {
			return nil, fmt.Errorf("unexpected response format: installed[%d].metadata.latestVersion is missing", index)
		}
		if app.Metadata.LatestDockerVersion == nil || strings.TrimSpace(*app.Metadata.LatestDockerVersion) == "" {
			return nil, fmt.Errorf("unexpected response format: installed[%d].metadata.latestDockerVersion is missing or empty", index)
		}

		if *app.App.Version < *app.Metadata.LatestVersion {
			updates = append(updates, AppUpdate{
				URN:            *app.Info.URN,
				CurrentVersion: *app.Info.Version,
				LatestVersion:  *app.Metadata.LatestDockerVersion,
				TipiVersion:    *app.App.Version,
				LatestTipi:     *app.Metadata.LatestVersion,
			})
		}
	}

	return updates, nil
}

func handleAPIResponse(spin *components.Spinner, resp *http.Response, err error, successMessage, errorMessage string) {
	if err != nil {
		spin.Fail(errorMessage)
		fmt.Printf("Error: %v\n", err)
		spin.Finish()
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if resp.StatusCode == 204 || resp.ContentLength == 0 {
			spin.Succeed(successMessage)
			spin.Finish()
			return
		}

		var body AppResponseBody
		err := json.NewDecoder(resp.Body).Decode(&body)

		if err != nil {
			spin.Fail("Failed to decode response body")
			fmt.Printf("Error decoding response: %v\n", err)
			spin.Finish()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		_, success := utils.WaitForEvent(ctx, time.Minute*2, func(event utils.EventData) bool {
			return event.RequestID == body.RequestID
		})

		if !success {
			spin.Fail(errorMessage)
		} else {
			spin.Succeed(successMessage)
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error code: %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		spin.Fail(errorMessage)
	}
}

func handleSimpleAPIResponse(spin *components.Spinner, resp *http.Response, err error, successMessage, errorMessage string) {
	if err != nil {
		spin.Fail(errorMessage)
		fmt.Printf("Error: %v\n", err)
		spin.Finish()
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		spin.Succeed(successMessage)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error code: %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		spin.Fail(errorMessage)
	}

	spin.Finish()
}

func handleAvailableUpdates() {
	spin := components.NewSpinner("Checking for available updates...")
	installedURL := utils.GetAPIBaseURL("apps/installed")
	resp, err := utils.APIRequest(installedURL, "GET", "")

	if err != nil {
		spin.Fail("Failed to check for available updates.")
		fmt.Printf("Error: %v\n", err)
		spin.Finish()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error code: %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		spin.Fail("Failed to check for available updates.")
		spin.Finish()
		return
	}

	var response InstalledAppsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		spin.Fail("Failed to parse API response.")
		fmt.Printf("Error: %v\n", err)
		spin.Finish()
		return
	}

	updates, err := FindAvailableUpdates(response)
	if err != nil {
		spin.Fail("Failed to validate API response.")
		fmt.Printf("Error: %v\n", err)
		spin.Finish()
		return
	}

	if len(updates) == 0 {
		spin.Succeed("All apps are up to date.")
		spin.Finish()
		return
	}

	spin.Succeed(fmt.Sprintf("Found %d app(s) with available updates.", len(updates)))
	spin.Finish()

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"App", "Docker Version", "Latest Docker", "Tipi Version", "Latest Tipi"})
	for _, u := range updates {
		table.Append([]string{
			u.URN,
			u.CurrentVersion,
			u.LatestVersion,
			fmt.Sprintf("%d", u.TipiVersion),
			fmt.Sprintf("%d", u.LatestTipi),
		})
	}
	table.Render()
}

func RunApp(args types.AppArgs) {
	lifecycleURL := utils.GetAPIBaseURL("app-lifecycle")
	backupsURL := utils.GetAPIBaseURL("backups")

	switch args.Command {
	case types.AppCommandStart:
		spin := components.NewSpinner(fmt.Sprintf("Starting app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/start", lifecycleURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := fmt.Sprintf("Failed to start app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App started successfully!", errorMessage)

	case types.AppCommandStop:
		spin := components.NewSpinner(fmt.Sprintf("Stopping app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/stop", lifecycleURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := fmt.Sprintf("Failed to stop app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App stopped successfully!", errorMessage)

	case types.AppCommandUninstall:
		spin := components.NewSpinner(fmt.Sprintf("Uninstalling app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/uninstall", lifecycleURL, args.ID)
		resp, err := utils.APIRequest(url, "DELETE", `{"removeBackups": false}`)
		errorMessage := fmt.Sprintf("Failed to uninstall app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App uninstalled successfully!", errorMessage)

	case types.AppCommandReset:
		spin := components.NewSpinner(fmt.Sprintf("Resetting app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/reset", lifecycleURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := fmt.Sprintf("Failed to reset app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App reset successfully!", errorMessage)

	case types.AppCommandBackup:
		spin := components.NewSpinner(fmt.Sprintf("Backing up app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/backup", backupsURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := fmt.Sprintf("Failed to backup app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App backup created successfully!", errorMessage)

	case types.AppCommandRestore:
		spin := components.NewSpinner(fmt.Sprintf("Restoring app %s from %s backup...", args.ID, args.BackupFilename))
		url := fmt.Sprintf("%s/%s/restore", backupsURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", fmt.Sprintf(`{"filename":"%s"}`, args.BackupFilename))
		errorMessage := fmt.Sprintf("Failed to restore app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App restored successfully!", errorMessage)

	case types.AppCommandListBackups:
		spin := components.NewSpinner(fmt.Sprintf("Getting backups for app %s...", args.ID))
		backups, err := utils.GetAppBackups(args.ID)
		if err != nil {
			spin.Fail(fmt.Sprintf("Failed to get backups for app %s.", args.ID))
			fmt.Printf("Error: %v\n", err)
			spin.Finish()
			return
		}
		if len(backups) == 0 {
			spin.Warn(fmt.Sprintf("No backups found for app %s.", args.ID))
			spin.Finish()
			return
		}
		spin.Succeed(fmt.Sprintf("Found %d backups for app %s.", len(backups), args.ID))
		spin.Finish()
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Backup Name", "Size", "Created At"})
		for _, backup := range backups {
			table.Append([]string{
				backup.Name,
				backup.Size,
				backup.CreatedAt,
			})
		}
		table.Render()

	case types.AppCommandDeleteBackup:
		spin := components.NewSpinner(fmt.Sprintf("Deleting backup %s...", args.BackupFilename))
		url := fmt.Sprintf("%s/%s", backupsURL, args.ID)
		resp, err := utils.APIRequest(url, "DELETE", fmt.Sprintf(`{"filename":"%s"}`, args.BackupFilename))
		errorMessage := fmt.Sprintf("Failed to delete backup %s. See logs/error.log for more details.", args.BackupFilename)
		handleAPIResponse(spin, resp, err, "Backup deleted successfully!", errorMessage)

	case types.AppCommandUpdate:
		spin := components.NewSpinner(fmt.Sprintf("Updating app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/update", lifecycleURL, args.ID)
		resp, err := utils.APIRequest(url, "PATCH", `{"performBackup": true}`)
		errorMessage := fmt.Sprintf("Failed to update app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App updated successfully!", errorMessage)

	case types.AppCommandAvailableUpdates:
		handleAvailableUpdates()

	case types.AppCommandStartAll:
		spin := components.NewSpinner("Starting all apps...")
		url := fmt.Sprintf("%s/start-all", lifecycleURL)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := "Failed to start all apps. See logs/error.log for more details."
		handleSimpleAPIResponse(spin, resp, err, "All apps start queued successfully!", errorMessage)

	case types.AppCommandStopAll:
		spin := components.NewSpinner("Stopping all apps...")
		url := fmt.Sprintf("%s/stop-all", lifecycleURL)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := "Failed to stop all apps. See logs/error.log for more details."
		handleSimpleAPIResponse(spin, resp, err, "All apps stop queued successfully!", errorMessage)
	}
}
