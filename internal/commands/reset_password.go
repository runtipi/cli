package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/utils"
	"golang.org/x/term"
)

type resetPasswordResponse struct {
	Success bool   `json:"success"`
	Email   string `json:"email"`
}

func promptNewPassword() (string, error) {
	fmt.Print("Enter your new password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	password := string(passwordBytes)
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}

	fmt.Print("Confirm your new password: ")
	confirmBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password confirmation: %w", err)
	}

	confirm := string(confirmBytes)
	if password != confirm {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func handleResetPasswordResponse(resp *http.Response, err error) {
	if err != nil {
		fmt.Printf("%s Unable to reset password: %v\n", color.RedString("✗"), err)
		return
	}

	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Printf("%s Unable to read API response: %v\n", color.RedString("✗"), readErr)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("%s Password reset failed (status %d).\n", color.RedString("✗"), resp.StatusCode)
		if len(body) > 0 {
			fmt.Printf("Response: %s\n", string(body))
		}
		return
	}

	var response resetPasswordResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("%s Unable to parse API response: %v\n", color.RedString("✗"), err)
		return
	}

	if !response.Success {
		fmt.Printf("%s Password reset did not succeed.\n", color.RedString("✗"))
		return
	}

	fmt.Printf("%s Password reset complete. You can now sign in as %s with your new password.\n", color.GreenString("✓"), response.Email)
}

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
		newPassword, promptErr := promptNewPassword()
		if promptErr != nil {
			fmt.Printf("%s %v\n", color.RedString("✗"), promptErr)
			return
		}

		baseURL := utils.GetAPIBaseURL("auth/reset-password")
		payload, marshalErr := json.Marshal(map[string]string{"newPassword": newPassword})
		if marshalErr != nil {
			fmt.Printf("%s Unable to encode password reset request: %v\n", color.RedString("✗"), marshalErr)
			return
		}

		resp, requestErr := utils.APIRequest(baseURL, http.MethodPost, string(payload))
		if requestErr != nil {
			fmt.Printf("%s Unable to reset password: %v\n", color.RedString("✗"), requestErr)
			return
		}

		handleResetPasswordResponse(resp, nil)
	} else {
		errorMsg := fmt.Sprintf(
			"%s Unable to create password reset request. You can manually create a file with `echo $(date +%%s) > %s` to initiate a password reset and retry this command. Error: %v",
			color.RedString("✗"),
			resetFilePath,
			err,
		)
		fmt.Println(errorMsg)
	}
}
