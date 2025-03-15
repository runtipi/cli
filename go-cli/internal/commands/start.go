package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/runtipi/cli/internal/components"
	"github.com/runtipi/cli/internal/types"
	"github.com/runtipi/cli/internal/utils"
)

// RunStart executes the start command
func RunStart(args types.StartArgs) {
	spin := components.NewSpinner("")

	// Check user permissions
	spin.SetMessage("Checking user permissions")
	if err := utils.EnsureDocker(); err != nil {
		spin.Fail(err.Error())
		spin.Finish()
		return
	}
	spin.Succeed("User permissions are ok")

	// Copy system files
	spin.SetMessage("Copying system files...")
	if err := utils.CopySystemFiles(); err != nil {
		spin.Fail("Failed to copy system files")
		spin.Finish()
		fmt.Printf("\nError: %v\n", err)
		return
	}
	spin.Succeed("Copied system files")

	// Generate env file
	spin.SetMessage("Generating .env file...")
	if err := utils.GenerateEnvFile(args.EnvFile); err != nil {
		spin.Fail("Failed to generate .env file")
		spin.Finish()
		fmt.Printf("\nError: %v\n", err)
		return
	}
	spin.Succeed("Generated .env file")

	// Ensure file permissions
	if !args.NoPermissions {
		spin.SetMessage("Ensuring file permissions... This may take a while depending on how many files there are to fix")
		if err := utils.EnsureFilePermissions(); err != nil {
			spin.Fail(err.Error())
			spin.Finish()
			return
		}
	}
	spin.Succeed("File permissions ok")

	// Pull images
	spin.SetMessage("Pulling images...")
	rootDir, err := os.Getwd()
	if err != nil {
		spin.Fail("Failed to get current directory")
		spin.Finish()
		fmt.Printf("\nError: %v\n", err)
		return
	}

	envFilePath := filepath.Join(rootDir, ".env")
	cmd := exec.Command("docker", "compose", "--env-file", envFilePath, "pull")
	if output, err := cmd.CombinedOutput(); err != nil {
		spin.Fail("Failed to pull images")
		spin.Finish()
		fmt.Printf("\nDebug: %s\n", output)
		return
	}
	spin.Succeed("Images pulled")

	// Stop and remove existing containers
	spin.SetMessage("Stopping existing containers...")
	containerNames := []string{
		// Legacy naming
		"tipi-reverse-proxy",
		"tipi-docker-proxy",
		"tipi-db",
		"tipi-redis",
		"tipi-worker",
		"tipi-dashboard",
		// New naming
		"runtipi",
		"runtipi-reverse-proxy",
		"runtipi-db",
		"runtipi-redis",
	}

	for _, container := range containerNames {
		exec.Command("docker", "stop", container).Run()
		exec.Command("docker", "rm", container).Run()
	}
	spin.Succeed("Existing containers stopped")

	// Start containers
	spin.SetMessage("Starting containers...")
	userComposeFile := filepath.Join(rootDir, "user-config", "tipi-compose.yml")
	dockerArgs := []string{
		"compose",
		"--project-name", "runtipi",
		"-f", filepath.Join(rootDir, "docker-compose.yml"),
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

	// Display success message
	internalIP := utils.GetEnvValue("INTERNAL_IP")
	if internalIP == "" {
		internalIP = "localhost"
	}

	nginxPort := utils.GetEnvValue("NGINX_PORT")
	if nginxPort == "" {
		nginxPort = "80"
	}

	ipAndPort := fmt.Sprintf("Visit http://%s:%s to access the dashboard", internalIP, nginxPort)
	boxTitle := "Runtipi started successfully"
	boxBody := fmt.Sprintf("%s\n\n%s\n\n%s",
		ipAndPort,
		"Find documentation and guides at: https://runtipi.io",
		"Runtipi is entirely written in TypeScript and we are looking for contributors!",
	)

	consoleBox := components.NewConsoleBox(boxTitle, boxBody, 80, "green")
	consoleBox.Print()
}

