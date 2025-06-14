package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

type AppResponseBody struct {
	RequestId string `json:"requestId"`
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
			return event.RequestId == body.RequestId
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

func RunApp(args types.AppArgs) {
	baseURL := utils.GetAPIBaseURL("app-lifecycle")

	switch args.Command {
	case types.AppCommandStart:
		spin := components.NewSpinner(fmt.Sprintf("Starting app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/start", baseURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := fmt.Sprintf("Failed to start app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App started successfully!", errorMessage)

	case types.AppCommandStop:
		spin := components.NewSpinner(fmt.Sprintf("Stopping app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/stop", baseURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := fmt.Sprintf("Failed to stop app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App stopped successfully!", errorMessage)

	case types.AppCommandUninstall:
		spin := components.NewSpinner(fmt.Sprintf("Uninstalling app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/uninstall", baseURL, args.ID)
		resp, err := utils.APIRequest(url, "DELETE", `{"removeBackups": false}`)
		errorMessage := fmt.Sprintf("Failed to uninstall app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App uninstalled successfully!", errorMessage)

	case types.AppCommandReset:
		spin := components.NewSpinner(fmt.Sprintf("Resetting app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/reset", baseURL, args.ID)
		resp, err := utils.APIRequest(url, "POST", "{}")
		errorMessage := fmt.Sprintf("Failed to reset app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App reset successfully!", errorMessage)

	case types.AppCommandUpdate:
		spin := components.NewSpinner(fmt.Sprintf("Updating app %s...", args.ID))
		url := fmt.Sprintf("%s/%s/update", baseURL, args.ID)
		resp, err := utils.APIRequest(url, "PATCH", `{"performBackup": true}`)
		errorMessage := fmt.Sprintf("Failed to update app %s. See logs/error.log for more details.", args.ID)
		handleAPIResponse(spin, resp, err, "App updated successfully!", errorMessage)

	case types.AppCommandStartAll:
		fmt.Println("Start all apps: Not implemented yet")
	}
}
