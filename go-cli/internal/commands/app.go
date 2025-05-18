package commands

import (
	"fmt"
	"io"
	"net/http"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

func handleAPIResponse(spin *components.Spinner, resp *http.Response, err error, successMessage, errorMessage string) {
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
	envMap := utils.GetEnvMap()

	internalIP := "localhost"
	if ip, ok := envMap["INTERNAL_IP"]; ok && ip != "" {
		internalIP = ip
	}

	nginxPort := utils.DefaultNginxPort
	if port, ok := envMap["NGINX_PORT"]; ok && port != "" {
		nginxPort = port
	}

	baseURL := fmt.Sprintf("http://%s:%s/api/app-lifecycle", internalIP, nginxPort)

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
