package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/runtipi/cli/internal/assets"
)

func hasCgroupMemory(cgroupPath string) bool {
	for _, files := range [][2]string{
		{"memory.current", "memory.stat"},
		{"memory/memory.usage_in_bytes", "memory/memory.stat"},
	} {
		if _, err := os.ReadFile(filepath.Join(cgroupPath, files[0])); err != nil {
			continue
		}
		if _, err := os.ReadFile(filepath.Join(cgroupPath, files[1])); err == nil {
			return true
		}
	}

	return false
}

func renderDockerCompose(cgroupPath string) ([]byte, error) {
	composeTemplate, err := template.New("docker-compose").Parse(string(assets.DockerComposeTemplate))
	if err != nil {
		return nil, err
	}

	options := struct {
		MountHostCgroup bool
	}{MountHostCgroup: hasCgroupMemory(cgroupPath)}
	var rendered bytes.Buffer

	if err := composeTemplate.Execute(&rendered, options); err != nil {
		return nil, err
	}

	return rendered.Bytes(), nil
}
