package types

type StartArgs struct {
	EnvFile       string
	NoPermissions bool
}

type UpdateArgs struct {
	Version       *VersionType
	EnvFile       string
	NoPermissions bool
}

type AppArgs struct {
	Command AppCommand
	ID      string
}

// AppCommand represents the subcommands available for the app command
type AppCommand string

const (
	AppCommandStart     AppCommand = "start"
	AppCommandStop      AppCommand = "stop"
	AppCommandUninstall AppCommand = "uninstall"
	AppCommandReset     AppCommand = "reset"
	AppCommandUpdate    AppCommand = "update"
	AppCommandStartAll  AppCommand = "start-all"
)

