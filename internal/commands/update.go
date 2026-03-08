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

const PreReleaseWarning = "You are updating to pre-release version which may contain bugs, we are not responsible for any issues that may arise"

func RunUpdate(args types.UpdateArgs) {
	spin := components.NewSpinner("")
	spin.SetMessage("Grabbing releases from GitHub")

	var wantedVersion string

	switch args.Version.String() {
	case "latest":
		latest, err := utils.GetReleases("https://api.github.com/repos/runtipi/runtipi/releases/latest")
		if err != nil {
			spin.Fail("Failed to fetch latest release")
			spin.Finish()
			fmt.Printf("\nError: %v\n", err)
			return
		}
		if len(latest) == 0 {
			spin.Fail("Failed to fetch latest release")
			spin.Finish()
			fmt.Printf("\nError: No releases found\n")
			return
		}
		wantedVersion = latest[0].TagName
	case "nightly":
		spin.Warn(PreReleaseWarning)
		wantedVersion = "nightly"
	case "prerelease":
		spin.Warn(PreReleaseWarning)
		releases, err := utils.GetReleases("https://api.github.com/repos/runtipi/runtipi/releases")
		if err != nil {
			spin.Fail("Failed to fetch latest prerelease")
			spin.Finish()
			fmt.Printf("\nError: %v\n", err)
			return
		}
		filtered := utils.FilterNonPreReleases(releases)
		if len(filtered) == 0 {
			spin.Fail("Failed to fetch latest prerelease")
			spin.Finish()
			fmt.Printf("\nError: No prerelease releases found\n")
			return
		}
		wantedVersion = filtered[0].TagName
	default:
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

	spin.Succeed("Runtipi updated successfully. Starting new CLI")
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
