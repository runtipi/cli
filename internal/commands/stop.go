package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/config"
)

func RunStop() {
	spin := components.NewSpinner("Stopping containers...")

	cmd := exec.Command(
		"docker",
		"compose",
		"-f", filepath.Join(config.RootFolder, "docker-compose.yml"),
		"down",
		"--remove-orphans",
		"--rmi",
		"local",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		spin.Fail("Failed to stop containers. Please try to stop them manually")
		spin.Finish()
		fmt.Printf("\nDebug: %s\n", output)
		return
	}

	containerNames := []string{
		"runtipi",
		"runtipi-reverse-proxy",
		"runtipi-db",
		"runtipi-queue",
	}

	for _, container := range containerNames {
		exec.Command("docker", "stop", container).Run()
		exec.Command("docker", "rm", container).Run()
	}

	spin.Succeed("Runtipi successfully stopped")
	spin.Finish()
}
