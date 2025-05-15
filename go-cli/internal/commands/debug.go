package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/utils"
	"github.com/shirou/gopsutil/v3/mem"
)

func RunDebug() {
	fmt.Println("⚠️ Make sure you have started tipi before running this command")

	operatingSystem := runtime.GOOS
	architecture := runtime.GOARCH

	memInfo, _ := mem.VirtualMemory()
	memGB := float64(memInfo.Total) / 1024.0 / 1024.0 / 1024.0

	var osVersion string
	if operatingSystem == "darwin" {
		cmd := exec.Command("sw_vers", "-productVersion")
		output, err := cmd.Output()
		if err == nil {
			osVersion = strings.TrimSpace(string(output))
		} else {
			osVersion = "Unknown"
		}
	} else if operatingSystem == "linux" {
		data, err := os.ReadFile("/etc/os-release")
		if err == nil {
			for line := range strings.SplitSeq(string(data), "\n") {
				if strings.HasPrefix(line, "VERSION_ID=") {
					osVersion = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
					break
				}
			}
		}

		if osVersion == "" {
			osVersion = "Unknown"
		}
	} else if operatingSystem == "windows" {
		cmd := exec.Command("cmd", "/c", "ver")
		output, err := cmd.Output()
		if err == nil {
			osVersion = strings.TrimSpace(string(output))
		} else {
			osVersion = "Unknown"
		}
	} else {
		osVersion = "Unknown"
	}

	fmt.Printf("--- %s ---\n", color.BlueString("System information"))
	sysTable := tablewriter.NewWriter(os.Stdout)
	sysTable.Append([]string{"OS", operatingSystem})
	sysTable.Append([]string{"OS Version", osVersion})
	sysTable.Append([]string{"Memory (GB)", fmt.Sprintf("%.2f", memGB)})
	sysTable.Append([]string{"Architecture", architecture})
	sysTable.Render()

	configFile := filepath.Join("user-config", "tipi-compose.yml")
	configExists := "No"
	if _, err := os.Stat(configFile); err == nil {
		configExists = color.YellowString("Yes")
	}

	fmt.Printf("\n--- %s ---\n", color.BlueString("Tipi configuration"))
	configTable := tablewriter.NewWriter(os.Stdout)
	configTable.Append([]string{"Custom tipi docker config", configExists})
	configTable.Render()

	fmt.Printf("\n--- %s ---\n", color.BlueString("Settings.json"))
	settingsFilePath := filepath.Join(config.RootFolder, "state", "settings.json")

	jsonData, err := os.ReadFile(settingsFilePath)
	if err == nil {
		var prettyJSON map[string]any
		err = json.Unmarshal(jsonData, &prettyJSON)
		if err == nil {
			prettyData, err := json.MarshalIndent(prettyJSON, "", "  ")
			if err == nil {
				fmt.Println(string(prettyData))
			}
		}
	}

	envMap := utils.GetEnvMap()
	fmt.Printf("\n--- %s ---\n", color.BlueString("Environment variables"))
	envTable := tablewriter.NewWriter(os.Stdout)

	getRedactedEnv := func(key string) string {
		if _, ok := envMap[key]; ok {
			return "<redacted>"
		}
		return color.RedString("Not set")
	}

	getEnvValue := func(key string) string {
		if val, ok := envMap[key]; ok {
			return val
		}
		return color.RedString("Not set")
	}

	envTable.Append([]string{"POSTGRES_PASSWORD", getRedactedEnv("POSTGRES_PASSWORD")})
	envTable.Append([]string{"RABBITMQ_PASSWORD", getRedactedEnv("RABBITMQ_PASSWORD")})
	envTable.Append([]string{"APPS_REPO_ID", getEnvValue("APPS_REPO_ID")})
	envTable.Append([]string{"APPS_REPO_URL", getEnvValue("APPS_REPO_URL")})
	envTable.Append([]string{"TIPI_VERSION", getEnvValue("TIPI_VERSION")})
	envTable.Append([]string{"INTERNAL_IP", getEnvValue("INTERNAL_IP")})
	envTable.Append([]string{"ARCHITECTURE", getEnvValue("ARCHITECTURE")})
	envTable.Append([]string{"JWT_SECRET", getRedactedEnv("JWT_SECRET")})
	envTable.Append([]string{"ROOT_FOLDER_HOST", getEnvValue("ROOT_FOLDER_HOST")})
	envTable.Append([]string{"RUNTIPI_APP_DATA_PATH", getEnvValue("RUNTIPI_APP_DATA_PATH")})
	envTable.Append([]string{"NGINX_PORT", getEnvValue("NGINX_PORT")})
	envTable.Append([]string{"NGINX_PORT_SSL", getEnvValue("NGINX_PORT_SSL")})
	envTable.Append([]string{"DOMAIN", getRedactedEnv("DOMAIN")})
	envTable.Append([]string{"POSTGRES_HOST", getEnvValue("POSTGRES_HOST")})
	envTable.Append([]string{"POSTGRES_DBNAME", getEnvValue("POSTGRES_DBNAME")})
	envTable.Append([]string{"POSTGRES_USERNAME", getEnvValue("POSTGRES_USERNAME")})
	envTable.Append([]string{"POSTGRES_PORT", getEnvValue("POSTGRES_PORT")})
	envTable.Append([]string{"RABBITMQ_HOST", getEnvValue("RABBITMQ_HOST")})
	envTable.Append([]string{"RABBITMQ_USERNAME", getEnvValue("RABBITMQ_USERNAME")})
	envTable.Append([]string{"DEMO_MODE", getEnvValue("DEMO_MODE")})
	envTable.Append([]string{"LOCAL_DOMAIN", getEnvValue("LOCAL_DOMAIN")})

	envTable.Render()

	fmt.Printf("\n--- %s ---\n", color.BlueString("Docker containers"))
	dockerTable := tablewriter.NewWriter(os.Stdout)

	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=runtipi", "--format", "{{.Names}} {{.Status}}")
	output, err := cmd.Output()

	if err == nil {
		containers := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(containers) > 0 && containers[0] != "" {
			for _, container := range containers {
				parts := strings.SplitN(container, " ", 2)
				if len(parts) >= 2 {
					name := parts[0]
					status := parts[1]

					statusColored := status
					if strings.Contains(status, "Up") {
						statusColored = color.GreenString(status)
					} else {
						statusColored = color.RedString(status)
					}

					dockerTable.Append([]string{name, statusColored})
				}
			}
		} else {
			dockerTable.Append([]string{"No containers found", ""})
		}
	} else {
		dockerTable.Append([]string{"No containers found", ""})
	}

	dockerTable.Render()

	fmt.Println("^ If a container is not 'Up', you can run the command `docker logs <container_name>` to see the logs of that container.")
}
