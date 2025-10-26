package commands

import (
	"fmt"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

func PrepareEnvironment(args types.StartArgs) error {
	spin := components.NewSpinner("Checking user permissions")
	if err := utils.EnsureDocker(); err != nil {
		spin.Fail(err.Error())
		spin.Finish()
		return err
	}
	spin.Succeed("User permissions are ok")

	spin.SetMessage("Copying system files...")
	if err := utils.CopySystemFiles(); err != nil {
		spin.Fail("Failed to copy system files")
		spin.Finish()
		return fmt.Errorf("failed to copy system files: %w", err)
	}
	spin.Succeed("Copied system files")

	spin.SetMessage("Generating .env file...")
	if err := utils.GenerateEnvFile(args.EnvFile); err != nil {
		spin.Fail("Failed to generate .env file")
		spin.Finish()
		return fmt.Errorf("failed to generate .env file: %w", err)
	}
	spin.Succeed("Generated .env file")

	if !args.NoPermissions {
		spin.SetMessage("Ensuring file permissions... This may take a while depending on how many files there are to fix")
		if err := utils.EnsureFilePermissions(); err != nil {
			spin.Fail(err.Error())
			spin.Finish()
			return fmt.Errorf("failed to ensure file permissions: %w", err)
		}
		spin.Succeed("File permissions ok")
	} else {
		spin.Warn("File permissions check skipped")
	}
	spin.Finish()

	return nil
}

func RunPrepare(args types.StartArgs) {
	if err := validateStartArgs(args); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := PrepareEnvironment(args); err != nil {
		fmt.Printf("\nError: %v\n", err)
		return
	}

	fmt.Println()
	boxTitle := "Environment prepared successfully"
	boxBody := "The environment has been prepared.\n\nYou can now run 'start' to launch the containers."

	consoleBox := components.NewConsoleBox(boxTitle, boxBody, 80, "Green")
	consoleBox.Print()
}
