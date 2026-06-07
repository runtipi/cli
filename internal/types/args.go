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
	Command        AppCommand
	ID             string
	BackupFilename string // Needed only for restore and delete backup commands it can be ommitted for other commands
	InstallOptions map[string]any
}

type AppInstallFlags struct {
	Port                    int
	MaxBackups              int
	Domain                  string
	LocalSubdomain          string
	Exposed                 bool
	ExposedLocal            bool
	OpenPort                bool
	VisibleOnGuestDashboard bool
	EnableAuth              bool
	SkipEnv                 bool
	SkipPull                bool
	SkipRun                 bool
	ForcePull               bool
	SetOptions              []string
}

// AppCommand represents the subcommands available for the app command
type AppCommand string

const (
	AppCommandStart        AppCommand = "start"
	AppCommandStop         AppCommand = "stop"
	AppCommandInstall      AppCommand = "install"
	AppCommandUninstall    AppCommand = "uninstall"
	AppCommandReset        AppCommand = "reset"
	AppCommandUpdate       AppCommand = "update"
	AppCommandBackup       AppCommand = "backup"
	AppCommandRestore      AppCommand = "restore"
	AppCommandListBackups  AppCommand = "list-backups"
	AppCommandDeleteBackup AppCommand = "delete-backup"
	AppCommandStartAll     AppCommand = "start-all"
	AppCommandStopAll      AppCommand = "stop-all"
)

type AppStoreArgs struct {
	Command AppStoreCommand
	URL     string
	Name    string
}

// AppStoreCommand represents the subcommands available for the appstore command
type AppStoreCommand string

const (
	AppStoreCommandUpdate AppStoreCommand = "update"
	AppStoreCommandList   AppStoreCommand = "list"
	AppStoreCommandAdd    AppStoreCommand = "add"
	AppStoreCommandRemove AppStoreCommand = "remove"
)
