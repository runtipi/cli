// Package commands
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

type AppResponseBody struct {
	RequestID string `json:"requestId"`
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

func RunApp(args types.AppArgs) {
	lifecycleURL := utils.GetAPIBaseURL("app-lifecycle")
	backupsURL := utils.GetAPIBaseURL("backups")

	switch args.Command {
	case types.AppCommandInstall:
		spin := components.NewSpinner(fmt.Sprintf("Installing app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/install", lifecycleURL, args.ID)

		payloadBytes, err := json.Marshal(args.InstallOptions)
		if err != nil {
			spin.Fail("Failed to prepare install request payload.")
			fmt.Printf("Error: %v\n", err)
			spin.Finish()
			return
		}

		resp, err := utils.APIRequest(url, "POST", string(payloadBytes))
		errorMessage := fmt.Sprintf("Failed to install app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App installed successfully!", errorMessage)

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
