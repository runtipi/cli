package commands

import (
	"fmt"

	"github.com/runtipi/cli/internal/utils"
)

func RunVersion() {
	version := utils.GetEnvValue("TIPI_VERSION")

	if version == "" {
		version = "unknown"
	}

	fmt.Println(version)
}
