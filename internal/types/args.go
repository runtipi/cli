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
	BackupFilename string
}


type AppCommand string

const (
	AppCommandStart        AppCommand = "start"
	AppCommandStop         AppCommand = "stop"
	AppCommandUninstall    AppCommand = "uninstall"
	AppCommandReset        AppCommand = "reset"
	AppCommandUpdate       AppCommand = "update"
	AppCommandBackup       AppCommand = "backup"
	AppCommandRestore      AppCommand = "restore"
	AppCommandListBackups  AppCommand = "list-backups"
	AppCommandDeleteBackup AppCommand = "delete-backup"
	AppCommandStartAll     AppCommand = "start-all"
)

type RepoArgs struct {
	Command RepoCommand
	URL     string
	Name    string
}

type RepoCommand string

const (
	RepoCommandUpdate RepoCommand = "update"
	RepoCommandList   RepoCommand = "list"
	RepoCommandAdd    RepoCommand = "add"
	RepoCommandRemove RepoCommand = "remove"
)
