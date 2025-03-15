package utils

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

type EnvMap map[string]string

func GetEnvMap() EnvMap {
	rootDir, err := os.Getwd()
	if err != nil {
		return make(EnvMap)
	}
	envFilePath := filepath.Join(rootDir, ".env")

	content, err := os.ReadFile(envFilePath)
	if err != nil {
		return make(EnvMap)
	}
	return EnvStringToMap(string(content))
}

func GetEnvValue(key string) string {
	envMap := GetEnvMap()
	return envMap[key]
}

func EnvStringToMap(envString string) EnvMap {
	envMap := make(EnvMap)

	for line := range strings.SplitSeq(envString, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envMap[key] = value
		}
	}

	return envMap
}

func EnvMapToString(envMap EnvMap) string {
	var lines []string
	for key, value := range envMap {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(lines, "\n") + "\n"
}

func GenerateEnvFile(customEnvFile string) error {
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get current directory: %w", err)
	}

	// Create required directories and files
	statePath := filepath.Join(rootDir, "state")
	settingsPath := filepath.Join(statePath, "settings.json")
	envFilePath := filepath.Join(rootDir, ".env")

	if err := os.MkdirAll(statePath, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Create empty .env file if it doesn't exist
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		if err := os.WriteFile(envFilePath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .env file: %w", err)
		}
	}

	// Create empty settings.json if it doesn't exist
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		if err := os.WriteFile(settingsPath, []byte("{}"), 0644); err != nil {
			return fmt.Errorf("failed to create settings.json: %w", err)
		}
	}

	if err := GenerateSeed(rootDir); err != nil {
		return fmt.Errorf("failed to generate seed: %w", err)
	}

	// Read existing env file
	currentEnv, err := os.ReadFile(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}
	envMap := EnvStringToMap(string(currentEnv))

	// Read settings.json
	settingsContent, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings.json: %w", err)
	}

	var settings SettingsSchema
	if err := json.Unmarshal(settingsContent, &settings); err != nil {
		return fmt.Errorf("failed to parse settings.json: %w", err)
	}

	// Read version
	version, err := os.ReadFile(filepath.Join(rootDir, "VERSION"))
	if err != nil {
		version = []byte("dev")
	}

	// Get seed
	seed, err := GetSeed(rootDir)
	if err != nil {
		return fmt.Errorf("failed to get seed: %w", err)
	}

	// Generate passwords if they don't exist
	postgresPassword := envMap["POSTGRES_PASSWORD"]
	if postgresPassword == "" {
		postgresPassword = DeriveEntropy("postgres_password", seed)
	}

	redisPassword := envMap["REDIS_PASSWORD"]
	if redisPassword == "" {
		redisPassword = DeriveEntropy("redis_password", seed)
	}

	// Handle app data path
	appDataPath := settings.AppDataPath
	if appDataPath == "" {
		appDataPath = settings.StoragePath
	}
	if appDataPath == "" {
		appDataPath = rootDir
	}

	// Validate app data path
	if appDataPath != rootDir {
		if _, err := os.Stat(appDataPath); os.IsNotExist(err) {
			return fmt.Errorf("path '%s' does not exist on your system. Make sure it is an absolute path or remove it from settings.json", appDataPath)
		}
	}

	// Create new environment map
	newEnv := EnvMap{
		"INTERNAL_IP":           settings.InternalIP,
		"ARCHITECTURE":          GetArchitecture(),
		"TIPI_VERSION":          string(version),
		"ROOT_FOLDER_HOST":      rootDir,
		"NGINX_PORT":            RawToString(settings.NginxPort),
		"NGINX_PORT_SSL":        RawToString(settings.NginxSSLPort),
		"RUNTIPI_APP_DATA_PATH": appDataPath,
		"POSTGRES_HOST":         "runtipi-db",
		"POSTGRES_PORT":         RawToString(settings.PostgresPort),
		"POSTGRES_DBNAME":       "tipi",
		"POSTGRES_USERNAME":     "tipi",
		"POSTGRES_PASSWORD":     postgresPassword,
		"REDIS_HOST":            "runtipi-redis",
		"REDIS_PASSWORD":        redisPassword,
		"DOMAIN":                settings.Domain,
		"LOCAL_DOMAIN":          settings.LocalDomain,
	}

	// Set default values if not present
	if newEnv["INTERNAL_IP"] == "" {
		newEnv["INTERNAL_IP"] = GetInternalIP()
	}
	if newEnv["NGINX_PORT"] == "" {
		newEnv["NGINX_PORT"] = DefaultNginxPort
	}
	if newEnv["NGINX_PORT_SSL"] == "" {
		newEnv["NGINX_PORT_SSL"] = DefaultNginxPortSSL
	}
	if newEnv["POSTGRES_PORT"] == "" {
		newEnv["POSTGRES_PORT"] = DefaultPostgresPort
	}
	if newEnv["DOMAIN"] == "" {
		newEnv["DOMAIN"] = DefaultDomain
	}
	if newEnv["LOCAL_DOMAIN"] == "" {
		newEnv["LOCAL_DOMAIN"] = DefaultLocalDomain
	}

	// Handle custom env file if provided
	if customEnvFile != "" {
		customEnvContent, err := os.ReadFile(customEnvFile)
		if err != nil {
			return fmt.Errorf("failed to read custom env file: %w", err)
		}

		customEnvMap := EnvStringToMap(string(customEnvContent))
		maps.Copy(newEnv, customEnvMap)
	}

	// Write the new env file
	envContent := EnvMapToString(newEnv)
	if err := os.WriteFile(envFilePath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}
