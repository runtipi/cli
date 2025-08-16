package main

import (
	"fmt"
	"os"

	"github.com/runtipi/cli/internal/commands"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/types"

	"github.com/spf13/cobra"
)

var (
	version   string
	commit    string
	buildDate string
)

func init() {
	var err error

	config.RootFolder = os.Getenv("ROOT_FOLDER_HOST")

	if config.RootFolder == "" {
		config.RootFolder, err = os.Getwd()
		if err != nil {
			fmt.Println("Error getting working directory:", err)
			os.Exit(1)
		}
	}

	if envRootFolder := os.Getenv("ROOT_FOLDER_HOST"); envRootFolder != "" {
		config.RootFolder = envRootFolder
	}

	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	if buildDate == "" {
		buildDate = "unknown"
	}

	config.Info = config.AppInfo{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "./runtipi-cli",
		Short: "Runtipi CLI tool",
		Long:  `Runtipi is a home server manager that helps you self-host your services easily.`,
	}

	// Common flags
	var startArgs types.StartArgs
	var restartArgs types.StartArgs
	var updateArgs types.UpdateArgs
	var appArgs types.AppArgs

	// Start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start Runtipi",
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunStart(startArgs)
		},
	}
	startCmd.Flags().StringVar(&startArgs.EnvFile, "env-file", "", "Path to a custom .env file. Can be relative to the current directory or absolute.")
	startCmd.Flags().BoolVar(&startArgs.NoPermissions, "no-permissions", false, "Skip setting file permissions (not recommended)")

	// Stop command
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop Runtipi",
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunStop()
		},
	}

	// Restart command
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart Runtipi",
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunStop()
			commands.RunStart(restartArgs)
		},
	}
	restartCmd.Flags().StringVar(&restartArgs.EnvFile, "env-file", "", "Path to a custom .env file. Can be relative to the current directory or absolute.")
	restartCmd.Flags().BoolVar(&restartArgs.NoPermissions, "no-permissions", false, "Skip setting file permissions (not recommended)")

	// Update command
	updateCmd := &cobra.Command{
		Use:   "update [version]",
		Short: "Update Runtipi",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			version, err := types.NewVersion(args[0])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			updateArgs.Version = version
			commands.RunUpdate(updateArgs)
		},
	}
	updateCmd.Flags().StringVar(&updateArgs.EnvFile, "env-file", "", "Path to a custom .env file. Can be relative to the current directory or absolute.")
	updateCmd.Flags().BoolVar(&updateArgs.NoPermissions, "no-permissions", false, "Skip setting file permissions (not recommended)")

	// Reset password command
	resetPasswordCmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset Runtipi password",
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunResetPassword()
		},
	}

	// Debug command
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug Runtipi",
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunDebug()
		},
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show Runtipi version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", config.Info.Version, config.Info.Commit, config.Info.BuildDate)
		},
	}

	// Installed apps command
	installedCmd := &cobra.Command{
		Use:   "installed",
		Short: "List installed apps",
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunInstalled()
		},
	}

	// App command and subcommands
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Manage Runtipi apps",
	}

	appStartCmd := &cobra.Command{
		Use:   "start [app-id]",
		Short: "Start an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandStart
			appArgs.ID = args[0]
			commands.RunApp(appArgs)
		},
	}

	appStopCmd := &cobra.Command{
		Use:   "stop [app-id]",
		Short: "Stop an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandStop
			appArgs.ID = args[0]
			commands.RunApp(appArgs)
		},
	}

	appUninstallCmd := &cobra.Command{
		Use:   "uninstall [app-id]",
		Short: "Uninstall an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandUninstall
			appArgs.ID = args[0]
			commands.RunApp(appArgs)
		},
	}

	appResetCmd := &cobra.Command{
		Use:   "reset [app-id]",
		Short: "Reset an app",
		Long:  "Reset an app to its initial state. This will delete all data and settings for the app.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandReset
			appArgs.ID = args[0]
			commands.RunApp(appArgs)
		},
	}

	appUpdateCmd := &cobra.Command{
		Use:   "update [app-id]",
		Short: "Update an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandUpdate
			appArgs.ID = args[0]
			commands.RunApp(appArgs)
		},
	}

	appBackupCmd := &cobra.Command{
		Use:   "backup [app-id]",
		Short: "Backup an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandBackup
			appArgs.ID = args[0]
			commands.RunApp(appArgs)
		},
	}

	appRestoreCmd := &cobra.Command{
		Use:   "restore [app-id] [backup-filename]",
		Short: "Restore an app from a backup",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandRestore
			appArgs.ID = args[0]
			appArgs.BackupFilename = args[1]
			commands.RunApp(appArgs)
		},
	}

	appListBackupsCmd := &cobra.Command{
		Use:   "list-backups [app-id]",
		Short: "List backups for an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandListBackups
			appArgs.ID = args[0]
			commands.RunApp(appArgs)
		},
	}

	appDeleteBackupCmd := &cobra.Command{
		Use:   "delete-backup [app-id] [backup-filename]",
		Short: "Delete a specific backup of an app",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandDeleteBackup
			appArgs.ID = args[0]
			appArgs.BackupFilename = args[1]
			commands.RunApp(appArgs)
		},
	}

	appStartAllCmd := &cobra.Command{
		Use:   "start-all",
		Short: "Start all apps",
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandStartAll
			commands.RunApp(appArgs)
		},
	}

	appCmd.AddCommand(appStartCmd)
	appCmd.AddCommand(appStopCmd)
	appCmd.AddCommand(appUninstallCmd)
	appCmd.AddCommand(appResetCmd)
	appCmd.AddCommand(appUpdateCmd)
	appCmd.AddCommand(appBackupCmd)
	appCmd.AddCommand(appRestoreCmd)
	appCmd.AddCommand(appListBackupsCmd)
	appCmd.AddCommand(appDeleteBackupCmd)
	appCmd.AddCommand(appStartAllCmd)

	// Add commands to root command
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(resetPasswordCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(appCmd)
	rootCmd.AddCommand(installedCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
