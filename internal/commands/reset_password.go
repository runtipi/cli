package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/utils"
)

func RunResetPassword() {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	statePath := filepath.Join(config.RootFolder, "state")
	if err := os.MkdirAll(statePath, 0755); err != nil {
		fmt.Printf("%s Unable to create state directory: %v\n", color.RedString("✗"), err)
		return
	}

	resetFilePath := filepath.Join(statePath, "password-change-request")
	err := os.WriteFile(resetFilePath, []byte(timestamp), 0644)

	if err == nil {
		internalIP := utils.GetEnvValue("INTERNAL_IP")
		if internalIP == "" {
			internalIP = "localhost"
		}

		nginxPort := utils.GetEnvValue("NGINX_PORT")
		if nginxPort == "" {
			nginxPort = "80"
		}

		resetURL := fmt.Sprintf("http://%s:%s/reset-password", internalIP, nginxPort)
		successMsg := fmt.Sprintf("%s Password reset request created. Head back to %s to set your new password.",
			color.GreenString("✓"), resetURL)

		fmt.Println(successMsg)
	} else {
		errorMsg := fmt.Sprintf(
			"%s Unable to create password reset request. You can manually create a file with `echo $(date +%%s) > %s` to initiate a password reset. Error: %v",
			color.RedString("✗"),
			resetFilePath,
			err,
		)
		fmt.Println(errorMsg)
	}
}
