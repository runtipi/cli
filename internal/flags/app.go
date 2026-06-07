package flags

import (
	"fmt"
	"strings"

	"github.com/runtipi/cli/internal/types"
	"github.com/spf13/cobra"
)


func BuildAppInstallOptions(cmd *cobra.Command, args types.AppInstallFlags) (map[string]any, error) {
	options := map[string]any{}
	commandFlags := cmd.Flags()

	if commandFlags.Changed("port") {
		options["port"] = args.Port
	}
	if commandFlags.Changed("max-backups") {
		options["maxBackups"] = args.MaxBackups
	}
	if commandFlags.Changed("domain") {
		options["domain"] = args.Domain
	}
	if commandFlags.Changed("local-subdomain") {
		options["localSubdomain"] = args.LocalSubdomain
	}
	if commandFlags.Changed("exposed") {
		options["exposed"] = args.Exposed
	}
	if commandFlags.Changed("exposed-local") {
		options["exposedLocal"] = args.ExposedLocal
	}
	if commandFlags.Changed("open-port") {
		options["openPort"] = args.OpenPort
	}
	if commandFlags.Changed("visible-on-guest-dashboard") {
		options["isVisibleOnGuestDashboard"] = args.VisibleOnGuestDashboard
	}
	if commandFlags.Changed("enable-auth") {
		options["enableAuth"] = args.EnableAuth
	}
	if commandFlags.Changed("skip-env") {
		options["skipEnv"] = args.SkipEnv
	}
	if commandFlags.Changed("skip-pull") {
		options["skipPull"] = args.SkipPull
	}
	if commandFlags.Changed("skip-run") {
		options["skipRun"] = args.SkipRun
	}
	if commandFlags.Changed("force-pull") {
		options["forcePull"] = args.ForcePull
	}

	for _, option := range args.SetOptions {
		key, value, ok := strings.Cut(option, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid --set value %q. Expected key=value", option)
		}
		options[key] = value
	}

	return options, nil
}

func BindAppInstallFlags(cmd *cobra.Command, args *types.AppInstallFlags) {
	cmd.Flags().IntVar(&args.Port, "port", 0, "Port to expose for the app")
	cmd.Flags().IntVar(&args.MaxBackups, "max-backups", 0, "Maximum backups to keep for the app")
	cmd.Flags().StringVar(&args.Domain, "domain", "", "Domain used when exposing the app on the internet")
	cmd.Flags().StringVar(&args.LocalSubdomain, "local-subdomain", "", "Local subdomain used when exposing the app locally")
	cmd.Flags().BoolVar(&args.Exposed, "exposed", false, "Expose the app on the internet")
	cmd.Flags().BoolVar(&args.ExposedLocal, "exposed-local", false, "Expose the app on the local network")
	cmd.Flags().BoolVar(&args.OpenPort, "open-port", true, "Open the app port")
	cmd.Flags().BoolVar(&args.VisibleOnGuestDashboard, "visible-on-guest-dashboard", false, "Display the app on the guest dashboard")
	cmd.Flags().BoolVar(&args.EnableAuth, "enable-auth", false, "Enable authentication for the app")
	cmd.Flags().BoolVar(&args.SkipEnv, "skip-env", false, "Skip app environment generation")
	cmd.Flags().BoolVar(&args.SkipPull, "skip-pull", false, "Skip pulling app images")
	cmd.Flags().BoolVar(&args.SkipRun, "skip-run", false, "Skip starting the app after installation")
	cmd.Flags().BoolVar(&args.ForcePull, "force-pull", false, "Force pulling app images")
	cmd.Flags().StringArrayVar(&args.SetOptions, "set", nil, "Set an app-specific install form value as key=value. Can be used multiple times")
}
