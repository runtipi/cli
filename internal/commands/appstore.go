package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorReset  = "\033[0m" // Reset to default color
)

// AppStore represents a single app store/repository
type AppStore struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Url     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// AppStoresResponse represents the API response for listing app stores
type AppStoresResponse struct {
	AppStores []AppStore `json:"appStores"`
}

// Since GO doeesn't allow for generic empty types, we define a simple struct with nothing in it
type EmptyResponse struct{}

func handleAppStoreAPIResponse[R any](spin *components.Spinner, resp *http.Response, err error, successMessage, errorMessage string, response R) (R, error) {
	// 1. Check for errors in the HTTP request
	if err != nil {
		spin.Fail(errorMessage)
		spin.Finish()
		return response, err
	}

	defer resp.Body.Close()

	// 2. Read response body
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		spin.Fail(errorMessage)
		spin.Finish()
		return response, readErr
	}

	// 3. Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		spin.Fail(errorMessage)
		spin.Finish()
		return response, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	// 4. Decode JSON Response into expected format
	if err := json.Unmarshal(body, &response); err != nil {
		spin.Fail(errorMessage)
		spin.Finish()
		return response, err
	}

	// 5. Everything fine
	spin.Succeed(successMessage)
	spin.Finish()
	return response, nil
}

func printAppStores(appStores []AppStore) {
	if len(appStores) == 0 {
		fmt.Println("No app stores found.")
		return
	}

	fmt.Printf("Found %d app stores:\n\n", len(appStores))
	for i, appStore := range appStores {
		fmt.Printf("%d. ", i+1)
		
		if appStore.Name != "" {
			fmt.Printf("── %s\n", appStore.Name)
		} else if appStore.Slug != "" {
			fmt.Printf("── %s\n", appStore.Slug)
		}
		
		if appStore.Url != "" {
			fmt.Printf("    ⎿ %s\n", appStore.Url)
		}
		
		if appStore.Enabled {
			fmt.Printf("    ⎿ "+colorGreen+"✓"+colorReset+" Enabled \n")
		} else {
			fmt.Printf("    ⎿ "+colorRed+"✗"+colorReset+" Disabled\n")
		}
		
		fmt.Println()
	}
}

func RunAppStore(args types.AppStoreArgs) {
	appStoresURL := utils.GetAPIBaseURL("marketplace")

	switch args.Command {
	case types.AppStoreCommandUpdate:
		spin := components.NewSpinner("Updating app stores...")
		url := fmt.Sprintf("%s/pull", appStoresURL)
		resp, err := utils.APIRequest(url, "POST", "{}")
		_, apiErr := handleAppStoreAPIResponse[EmptyResponse](spin, resp, err, 
			"App stores updated successfully!", 
			"Failed to update app stores.", 
			EmptyResponse{})

		if apiErr != nil {
			fmt.Printf("Error updating app stores: %v\n", apiErr)
		}

	case types.AppStoreCommandList:
		spin := components.NewSpinner("Retrieving app stores...")
		url := fmt.Sprintf("%s/all", appStoresURL)
		resp, err := utils.APIRequest(url, "GET", "")
		result, apiErr := handleAppStoreAPIResponse[AppStoresResponse](spin, resp, err, 
			"App stores retrieved successfully!",
			"Failed to retrieve app stores.",
			AppStoresResponse{})

		if apiErr != nil {
			fmt.Printf("Error retrieving app stores: %v\n", apiErr)
		} else {
			printAppStores(result.AppStores)
		}

	case types.AppStoreCommandAdd:
		if args.Name == "" || args.URL == "" {
			fmt.Println("✗ Error: Both name and URL are required for adding an app store.")
			fmt.Println("Usage: runtipi appstore add <name> <url>")
			return
		}
		
		spin := components.NewSpinner(fmt.Sprintf("Adding app store %s...", args.Name))
		url := fmt.Sprintf("%s/create", appStoresURL)
		
		// Use proper JSON marshaling for safety
		payload := map[string]string{
			"name": args.Name,
			"url":  args.URL,
		}
		payloadBytes, mErr := json.Marshal(payload)
		if mErr != nil {
			spin.Fail("Failed to prepare request payload.")
			fmt.Printf("Error: %v\n", mErr)
			spin.Finish()
			return
		}
		
		resp, err := utils.APIRequest(url, "POST", string(payloadBytes))
		_, apiErr := handleAppStoreAPIResponse[EmptyResponse](spin, resp, err,
			fmt.Sprintf("App store %s added successfully!", args.Name),
			fmt.Sprintf("Failed to add app store %s.", args.Name),
			EmptyResponse{})

		if apiErr != nil {
			fmt.Printf("Error adding app store: %v\n", apiErr)
		}

	case types.AppStoreCommandRemove:
		if args.Name == "" {
			fmt.Println("✗ Error: App store name is required for removal.")
			fmt.Println("Usage: runtipi appstore remove <name>")
			return
		}
		
		spin := components.NewSpinner(fmt.Sprintf("Removing app store %s...", args.Name))
		// Use URL escaping for safety
		escapedName := url.PathEscape(args.Name)
		url := fmt.Sprintf("%s/%s", appStoresURL, escapedName)
		resp, err := utils.APIRequest(url, "DELETE", "")
		handleAppStoreAPIResponse(spin, resp, err,
			fmt.Sprintf("App store %s removed successfully! (Note: Shows success even if app store doesn't exist - check Runtipi logs for actual status)", args.Name),
			fmt.Sprintf("Failed to remove app store %s.", args.Name),
			EmptyResponse{})

		// For now we cannot handle errors for the Remove Command since runtipi always returns 200 even if the appstore doesn't exist and the error is only shown in the console.
	}
}