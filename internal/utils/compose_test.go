package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderDockerComposeOptionalCgroupMount(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		wantMount bool
	}{
		{name: "missing cgroup directory"},
		{name: "no memory accounting", files: map[string]string{"cpu.stat": "usage_usec 1"}},
		{name: "incomplete v2 accounting", files: map[string]string{"memory.current": "100"}},
		{name: "incomplete v1 accounting", files: map[string]string{"memory/memory.usage_in_bytes": "100"}},
		{
			name: "v2 memory accounting",
			files: map[string]string{
				"memory.current": "100",
				"memory.stat":    "inactive_file 0",
			},
			wantMount: true,
		},
		{
			name: "v1 memory accounting",
			files: map[string]string{
				"memory/memory.usage_in_bytes": "100",
				"memory/memory.stat":           "total_inactive_file 0",
			},
			wantMount: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cgroupPath := filepath.Join(t.TempDir(), "cgroup")
			for name, content := range test.files {
				file := filepath.Join(cgroupPath, name)
				if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(file, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			rendered, err := renderDockerCompose(cgroupPath)
			if err != nil {
				t.Fatal(err)
			}
			compose := string(rendered)
			if got := strings.Contains(compose, "target: /host/sys/fs/cgroup"); got != test.wantMount {
				t.Fatalf("cgroup mount present = %t, want %t", got, test.wantMount)
			}
			if test.wantMount {
				mount := "      - type: bind\n        source: /sys/fs/cgroup\n        target: /host/sys/fs/cgroup\n        read_only: true\n        bind:\n          create_host_path: false\n"
				if !strings.Contains(compose, mount) {
					t.Fatal("cgroup mount must be read-only and must never create a host path")
				}
			}
			if strings.Contains(compose, "{{") {
				t.Fatal("generated Compose file contains unrendered template directives")
			}
			if !strings.Contains(compose, "      - /proc:/host/proc\n") || !strings.Contains(compose, "      - /etc/localtime:/etc/localtime:ro\n") {
				t.Fatal("rendering removed an existing host mount")
			}
		})
	}
}

func TestRenderDockerComposeOmitsUnreadableCgroupAccounting(t *testing.T) {
	cgroupPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(cgroupPath, "memory.current"), []byte("100"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(cgroupPath, "memory.stat"), 0755); err != nil {
		t.Fatal(err)
	}

	rendered, err := renderDockerCompose(cgroupPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rendered), "/host/sys/fs/cgroup") {
		t.Fatal("generated Compose file mounts cgroups despite an accounting read failure")
	}
}
