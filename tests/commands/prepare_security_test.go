package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runtipi/cli/internal/commands"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/types"
)

func TestNoPermissionsRetainsSharedAccess(t *testing.T) {
	previousRoot := config.RootFolder
	config.RootFolder = t.TempDir()
	t.Cleanup(func() { config.RootFolder = previousRoot })
	if err := os.Chmod(config.RootFolder, 0755); err != nil {
		t.Fatal(err)
	}

	for _, directory := range []string{"state", "apps", "user-config"} {
		directoryPath := filepath.Join(config.RootFolder, directory)
		if err := os.Mkdir(directoryPath, 0777); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(directoryPath, 0777); err != nil {
			t.Fatal(err)
		}
	}

	bin := t.TempDir()
	docker := "#!/bin/sh\nif [ \"$1\" = version ]; then echo 28.0.0; else echo 2.40.0; fi\n"
	if err := os.WriteFile(filepath.Join(bin, "docker"), []byte(docker), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := commands.PrepareEnvironment(types.StartArgs{NoPermissions: true}); err != nil {
		t.Fatal(err)
	}

	for file, expected := range map[string]os.FileMode{
		".": 0755, "state": 0777, "user-config": 0777, ".env": 0644, "state/seed": 0644, "state/settings.json": 0644,
	} {
		info, err := os.Stat(filepath.Join(config.RootFolder, file))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != expected {
			t.Errorf("%s mode = %o, want %o", file, info.Mode().Perm(), expected)
		}
	}

	if _, err := os.Stat(filepath.Join(config.RootFolder, ".secrets")); !os.IsNotExist(err) {
		t.Fatal("preparation created host authentication storage")
	}
}
