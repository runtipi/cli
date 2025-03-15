package main

import (
	"fmt"
	"os"

	"github.com/runtipi/cli/internal/commands"
	"github.com/runtipi/cli/internal/types"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "runtipi",
		Short: "Runtipi CLI tool",
		Long:  `Runtipi is a home server manager that helps you self-host your services easily.`,
	}

	// Common flags
	var startArgs types.StartArgs
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
	startCmd.Flags().StringVar(&startArgs.EnvFile, "env-file", "", "Path to a custom .env file")
	startCmd.Flags().BoolVar(&startArgs.NoPermissions, "no-permissions", false, "Skip setting file permissions (not recommended)")

	// Stop command
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop Runtipi",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement stop command
			fmt.Println("Stopping Runtipi...")
		},
	}

	// Restart command
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart Runtipi",
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunStart(startArgs)
		},
	}
	restartCmd.Flags().StringVar(&startArgs.EnvFile, "env-file", "", "Path to a custom .env file")
	restartCmd.Flags().BoolVar(&startArgs.NoPermissions, "no-permissions", false, "Skip setting file permissions (not recommended)")

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
			// TODO: Implement update command using updateArgs
			fmt.Printf("Updating Runtipi with args: %+v\n", updateArgs)
		},
	}
	updateCmd.Flags().StringVar(&updateArgs.EnvFile, "env-file", "", "Path to a custom .env file")
	updateCmd.Flags().BoolVar(&updateArgs.NoPermissions, "no-permissions", false, "Skip setting file permissions (not recommended)")

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
			fmt.Printf("Starting app: %s\n", appArgs.ID)
		},
	}

	appStopCmd := &cobra.Command{
		Use:   "stop [app-id]",
		Short: "Stop an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandStop
			appArgs.ID = args[0]
			fmt.Printf("Stopping app: %s\n", appArgs.ID)
		},
	}

	appUninstallCmd := &cobra.Command{
		Use:   "uninstall [app-id]",
		Short: "Uninstall an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandUninstall
			appArgs.ID = args[0]
			fmt.Printf("Uninstalling app: %s\n", appArgs.ID)
		},
	}

	appResetCmd := &cobra.Command{
		Use:   "reset [app-id]",
		Short: "Reset an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandReset
			appArgs.ID = args[0]
			fmt.Printf("Resetting app: %s\n", appArgs.ID)
		},
	}

	appUpdateCmd := &cobra.Command{
		Use:   "update [app-id]",
		Short: "Update an app",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandUpdate
			appArgs.ID = args[0]
			fmt.Printf("Updating app: %s\n", appArgs.ID)
		},
	}

	appStartAllCmd := &cobra.Command{
		Use:   "start-all",
		Short: "Start all apps",
		Run: func(cmd *cobra.Command, args []string) {
			appArgs.Command = types.AppCommandStartAll
			fmt.Println("Starting all apps...")
		},
	}

	appCmd.AddCommand(appStartCmd)
	appCmd.AddCommand(appStopCmd)
	appCmd.AddCommand(appUninstallCmd)
	appCmd.AddCommand(appResetCmd)
	appCmd.AddCommand(appUpdateCmd)
	appCmd.AddCommand(appStartAllCmd)

	// Reset password command
	resetPasswordCmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset Runtipi password",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement reset password command
			fmt.Println("Resetting password...")
		},
	}

	// Debug command
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug Runtipi",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement debug command
			fmt.Println("Running debug...")
		},
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show Runtipi version",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement version command
			fmt.Println("Runtipi version...")
		},
	}

	// Add commands to root command
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(resetPasswordCmd)
	rootCmd.AddCommand(appCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
