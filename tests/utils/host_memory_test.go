package utils_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/utils"
	"github.com/shirou/gopsutil/v3/mem"
)

func TestPreparationRefreshesHostMemory(t *testing.T) {
	previousRoot := config.RootFolder
	config.RootFolder = t.TempDir()
	t.Cleanup(func() { config.RootFolder = previousRoot })

	if err := utils.CopySystemFiles(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config.RootFolder, ".env"), []byte("RUNTIPI_HOST_MEMORY_TOTAL=1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := utils.GenerateEnvFile(""); err != nil {
		t.Fatal(err)
	}

	memory, err := mem.VirtualMemory()
	if err != nil {
		t.Fatal(err)
	}
	expected := strconv.FormatUint(memory.Total, 10)
	if total := utils.GetEnvValue("RUNTIPI_HOST_MEMORY_TOTAL"); total != expected {
		t.Fatalf("host memory total = %s, want %s", total, expected)
	}
}
