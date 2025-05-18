package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

func RunUpdate(args types.UpdateArgs) {
	spin := components.NewSpinner("")
	spin.SetMessage("Grabbing releases from GitHub")

	var wantedVersion string
	if args.Version.IsLatest() {
		latest, err := utils.GetLatestRelease()
		if err != nil {
			spin.Fail("Failed to fetch latest release")
			spin.Finish()
			fmt.Printf("\nError: %v\n", err)
			return
		}
		wantedVersion = latest
	} else if args.Version.IsNightly() {
		wantedVersion = "nightly"
	} else {
		wantedVersion = args.Version.String()
	}

	if utils.IsMajorBump(config.Info.Version, wantedVersion) {
		spin.Fail("You are trying to update to a new major version. Please update manually using the update instructions on the website. https://runtipi.io/docs/reference/breaking-updates")
		spin.Finish()
		return
	}

	err := utils.DownloadReleaseAndSelfReplace(wantedVersion)
	if err != nil {
		spin.Fail("Failed to download release")
		spin.Finish()
		fmt.Printf("\nError: %v\n", err)
		return
	}

	spin.Succeed("Tipi updated successfully. Starting new CLI")
	spin.Finish()

	newExecutablePath := filepath.Join(config.RootFolder, "runtipi-cli")

	var runArgs []string
	runArgs = append(runArgs, "start")

	if args.NoPermissions {
		runArgs = append(runArgs, "--no-permissions")
	}

	if args.EnvFile != "" {
		runArgs = append(runArgs, "--env-file", args.EnvFile)
	}

	cmd := exec.Command(newExecutablePath, runArgs...)
	cmd.Stdout = io.Writer(os.Stdout)
	cmd.Stderr = io.Writer(os.Stderr)
	cmd.Stdin = io.Reader(os.Stdin)

	err = cmd.Run()
	if err != nil {
		spin.Fail("Failed to start new CLI")
		fmt.Printf("\nDebug: %s\n", err)
		return
	}

	spin.Finish()
}
