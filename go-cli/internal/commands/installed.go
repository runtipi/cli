package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/utils"
)

func handleInstalledResponse(spin *components.Spinner, resp *http.Response, err error) {
	if err != nil {
		spin.Fail("Failed to retrieve installed apps.")
		fmt.Printf("Error: %v\n", err)
		spin.Finish()
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			spin.Fail("Failed to read API response.")
			fmt.Printf("Error: %v\n", readErr)
			spin.Finish()
			return
		}

		var response map[string]any
		if jsonErr := json.Unmarshal(body, &response); jsonErr != nil {
			spin.Fail("Failed to parse API response.")
			fmt.Printf("Error: %v\n", jsonErr)
			spin.Finish()
			return
		}

		installedApps, ok := response["installed"].([]any)
		if !ok {
			spin.Fail("Unexpected response format: 'installed' field is missing or invalid.")
			spin.Finish()
			return
		}

		if len(installedApps) == 0 {
			spin.Succeed("No apps found.")
		} else {
			spin.Finish()
			for _, app := range installedApps {
				appMap, ok := app.(map[string]any)
				if !ok {
					continue
				}

				info, ok := appMap["info"].(map[string]any)
				if !ok {
					continue
				}

				urn, ok := info["urn"].(string)
				if !ok {
					fmt.Println("Warning: Missing or invalid 'urn' field in app info.")
					continue
				}

				fmt.Println(urn)
			}
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error code: %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		spin.Fail("Failed to retrieve installed apps.")
	}

	spin.Finish()
}

func RunInstalled() {
	baseURL := utils.GetAPIBaseURL("apps/installed")

	spin := components.NewSpinner("Getting list of installed apps...")
	resp, err := utils.APIRequest(baseURL, "GET", "")
	handleInstalledResponse(spin, resp, err)
}
