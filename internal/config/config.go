// Package config provides application configuration details.
package config

type AppInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

var (
	RootFolder string
	Info       AppInfo
)
