package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

func validateStartArgs(args types.StartArgs) error {
	if args.EnvFile != "" {
		if _, err := os.Stat(args.EnvFile); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", args.EnvFile)
		}
	}
	return nil
}

func RunStart(args types.StartArgs) {
	if err := validateStartArgs(args); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := PrepareEnvironment(args); err != nil {
		fmt.Printf("\nError: %v\n", err)
		return
	}

	spin := components.NewSpinner("Pulling images...")

	envFilePath := filepath.Join(config.RootFolder, ".env")
	cmd := exec.Command(
		"docker",
		"compose",
		"--env-file",
		envFilePath,
		"-f", filepath.Join(config.RootFolder, "docker-compose.yml"),
		"pull")

	if output, err := cmd.CombinedOutput(); err != nil {
		spin.Fail("Failed to pull images")
		spin.Finish()
		fmt.Printf("\nDebug: %s\n", output)
		return
	}
	spin.Succeed("Images pulled")

	spin.SetMessage("Stopping existing containers...")
	containerNames := []string{
		"runtipi",
		"runtipi-reverse-proxy",
		"runtipi-db",
		"runtipi-redis",
		"runtipi-queue",
	}

	for _, container := range containerNames {
		exec.Command("docker", "stop", container).Run()
		exec.Command("docker", "rm", container).Run()
	}
	spin.Succeed("Existing containers stopped")

	spin.SetMessage("Starting containers...")
	userComposeFile := filepath.Join(config.RootFolder, "user-config", "tipi-compose.yml")
	dockerArgs := []string{
		"compose",
		"--project-name", "runtipi",
		"-f", filepath.Join(config.RootFolder, "docker-compose.yml"),
	}

	if _, err := os.Stat(userComposeFile); err == nil {
		dockerArgs = append(dockerArgs, "-f", userComposeFile)
	}

	dockerArgs = append(dockerArgs,
		"--env-file", envFilePath,
		"up",
		"--detach",
		"--remove-orphans",
		"--build",
	)

	cmd = exec.Command("docker", dockerArgs...)
	if output, err := cmd.CombinedOutput(); err != nil {
		spin.Fail("Failed to start containers")
		spin.Finish()
		fmt.Printf("\nDebug: %s\n", output)
		return
	}
	spin.Succeed("Containers started")
	spin.Finish()
	fmt.Println()

	internalIP := utils.GetEnvValue("INTERNAL_IP")
	if internalIP == "" {
		internalIP = "localhost"
	}

	nginxPort := utils.GetEnvValue("NGINX_PORT")
	if nginxPort == "" {
		nginxPort = "80"
	}

	ipAndPort := fmt.Sprintf("Visit http://%s:%s to access the dashboard", internalIP, nginxPort)
	boxTitle := "Runtipi started successfully 🎉"
	boxBody := fmt.Sprintf("%s\n\n%s\n\n%s",
		ipAndPort,
		"Find documentation and guides at https://runtipi.io",
		"We are looking for contributors!",
	)

	consoleBox := components.NewConsoleBox(boxTitle, boxBody, 80, "Green")
	consoleBox.Print()
}
