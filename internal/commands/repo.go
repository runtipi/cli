package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

func handleRepoAPIResponse(spin *components.Spinner, resp *http.Response, err error, successMessage, errorMessage string, customHandler func([]byte) error) {
	if err != nil {
		spin.Fail(errorMessage)
		fmt.Printf("Error: %v\n", err)
		spin.Finish()
		return
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		spin.Fail("Failed to read API response.")
		fmt.Printf("Error: %v\n", readErr)
		spin.Finish()
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if customHandler != nil {
			if handlerErr := customHandler(body); handlerErr != nil {
				spin.Fail(errorMessage)
				fmt.Printf("Error: %v\n", handlerErr)
				spin.Finish()
				return
			}
			spin.Succeed(successMessage)
			spin.Finish()
			return
		}

		if resp.StatusCode == 204 || len(body) == 0 {
			spin.Succeed(successMessage)
			spin.Finish()
			return
		}

		spin.Succeed(successMessage)
	} else {
		var errorResponse map[string]interface{}
		if jsonErr := json.Unmarshal(body, &errorResponse); jsonErr == nil {
			if message, exists := errorResponse["message"]; exists {
				fmt.Printf("API Error: %v\n", message)
				
				switch message {
				case "REPO_NOT_FOUND", "APP_STORE_NOT_FOUND":
					spin.Fail("Repository not found")
				case "SERVER_ERROR_DUPLICATE_APP_STORE_NAME":
					spin.Fail("Repository name already exists. Please choose a different name.")
				case "APP_STORE_CLONE_ERROR":
					if params, ok := errorResponse["intlParams"].(map[string]interface{}); ok {
						if url, ok := params["url"].(string); ok {
							spin.Fail(fmt.Sprintf("Failed to clone repository from URL: %s\nPlease verify that this is a valid Git repository accessible via HTTPS.", url))
						} else {
							spin.Fail("Failed to clone repository - invalid URL\nPlease verify that this is a valid Git repository accessible via HTTPS.")
						}
					} else {
						spin.Fail("Failed to clone repository - invalid URL\nPlease verify that this is a valid Git repository accessible via HTTPS.")
					}
				default:
					spin.Fail(fmt.Sprintf("%s: %v", errorMessage, message))
				}
			} else {
				spin.Fail(errorMessage)
			}
		} else {
			spin.Fail(errorMessage)
		}
	}

	spin.Finish()
}

func handleRepositoryListData(body []byte) error {
	var response map[string]any
	if jsonErr := json.Unmarshal(body, &response); jsonErr != nil {
		return fmt.Errorf("failed to parse API response: %v", jsonErr)
	}

	repositories, ok := response["appStores"].([]any)
	if !ok {
		if repoData, exists := response["repositories"].([]any); exists {
			repositories = repoData
		} else if repoArray, exists := response["data"].([]any); exists {
			repositories = repoArray
		} else if repoArray, exists := response["repos"].([]any); exists {
			repositories = repoArray
		} else {
			return fmt.Errorf("unexpected response format: repository data not found")
		}
	}

	if len(repositories) == 0 {
		fmt.Println("No repositories found.")
		return nil
	}

	fmt.Printf("Found %d app stores:\n\n", len(repositories))
	for i, repo := range repositories {
		repoMap, ok := repo.(map[string]any)
		if !ok {
			continue
		}
		slug := getStringField(repoMap, "slug", "id")
		name := getStringField(repoMap, "name", "title")
		url := getStringField(repoMap, "url", "git_url", "clone_url", "html_url")
		enabled := getBoolField(repoMap, "enabled")

		if name != "" || url != "" {
			fmt.Printf("%d. ", i+1)
			
			if name != "" {
				fmt.Printf("📦 %s", name)
			} else if slug != "" {
				fmt.Printf("📦 %s", slug)
			}
			
			if url != "" {
				fmt.Printf("\n   🔗 %s", url)
			}
			
			status := "❓ Unknown"
			if enabled {
				status = "✅ Enabled"
			} else {
				status = "❌ Disabled"
			}
			fmt.Printf("\n   %s\n", status)
			
			fmt.Println()
		}
	}
	return nil
}

func getStringField(data map[string]any, fieldNames ...string) string {
	for _, fieldName := range fieldNames {
		if value, ok := data[fieldName].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func getBoolField(data map[string]any, fieldName string) bool {
	if value, ok := data[fieldName].(bool); ok {
		return value
	}
	return false
}

func RunRepo(args types.RepoArgs) {
	reposURL := utils.GetAPIBaseURL("marketplace")

	switch args.Command {
	case types.RepoCommandUpdate:
		spin := components.NewSpinner("Updating repositories...")
		url := fmt.Sprintf("%s/pull", reposURL)
		resp, err := utils.APIRequest(url, "POST", "{}")
		handleRepoAPIResponse(spin, resp, err, 
			"Repositories updated successfully!", 
			"Failed to update repositories.", 
			nil)

	case types.RepoCommandList:
		spin := components.NewSpinner("Retrieving app repositories...")
		url := fmt.Sprintf("%s/all", reposURL)
		resp, err := utils.APIRequest(url, "GET", "")
		handleRepoAPIResponse(spin, resp, err, 
			"Repositories retrieved successfully!", 
			"Failed to retrieve app repositories.", 
			handleRepositoryListData)

	case types.RepoCommandAdd:
		if args.Name == "" || args.URL == "" {
			fmt.Println("✗ Error: Both name and URL are required for adding a repository.")
			fmt.Println("Usage: runtipi repo add <name> <url>")
			return
		}
		
		spin := components.NewSpinner(fmt.Sprintf("Adding repository %s...", args.Name))
		url := fmt.Sprintf("%s/create", reposURL)
		payload := fmt.Sprintf(`{"name":"%s","url":"%s"}`, args.Name, args.URL)
		resp, err := utils.APIRequest(url, "POST", payload)
		handleRepoAPIResponse(spin, resp, err,
			fmt.Sprintf("Repository %s added successfully!", args.Name),
			fmt.Sprintf("Failed to add repository %s.", args.Name),
			nil)

	case types.RepoCommandRemove:
		if args.Name == "" {
			fmt.Println("✗ Error: Repository name is required for removal.")
			fmt.Println("Usage: runtipi repo remove <name>")
			return
		}
		
		spin := components.NewSpinner(fmt.Sprintf("Removing repository %s...", args.Name))
		url := fmt.Sprintf("%s/%s", reposURL, args.Name)
		resp, err := utils.APIRequest(url, "DELETE", "")
		successMsg := fmt.Sprintf("Repository %s removed successfully! (Note: Shows success even if repository doesn't exist - check Runtipi logs for actual status)", args.Name)
		handleRepoAPIResponse(spin, resp, err,
			successMsg,
			fmt.Sprintf("Failed to remove repository %s.", args.Name),
			nil)
	}
}