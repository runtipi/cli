package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/config"
)

func RunStop() {
	spin := components.NewSpinner("Stopping containers...")

	if err := stopBackend("runtipi"); err != nil {
		spin.Fail("Failed to drain the backend")
		spin.Finish()
		fmt.Printf("\nError: %v\n", err)
		return
	}

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
		if container != "runtipi" {
			exec.Command("docker", "stop", container).Run()
		}
		exec.Command("docker", "rm", container).Run()
	}

	spin.Succeed("Runtipi successfully stopped")
	spin.Finish()
}

func stopBackend(containerName string) error {
	output, err := exec.Command("docker", "container", "ls", "--all", "--quiet", "--filter", "name=^/"+containerName+"$").CombinedOutput()
	if err != nil {
		return fmt.Errorf("find backend container: %w: %s", err, strings.TrimSpace(string(output)))
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return nil
	}

	if output, err := exec.Command("docker", "stop", "--timeout", "-1", containerID).CombinedOutput(); err != nil {
		return fmt.Errorf("stop backend container: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}
