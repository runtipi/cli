package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fakeShutdownDocker(t *testing.T, mode string) string {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "calls")
	script := `#!/bin/sh
echo "$*" >> "$SHUTDOWN_TEST_LOG"
if [ "$1" = container ]; then
  case "$SHUTDOWN_TEST_MODE" in
    absent) exit 0 ;;
    unavailable) echo 'daemon unavailable' >&2; exit 1 ;;
  esac
  echo backend-id
fi
if [ "$1" = stop ] && [ "$4" = backend-id ] && [ "$SHUTDOWN_TEST_MODE" = stop-failed ]; then
  echo 'stop failed' >&2
  exit 1
fi
`
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("SHUTDOWN_TEST_LOG", logPath)
	t.Setenv("SHUTDOWN_TEST_MODE", mode)

	return logPath
}

func TestRunStopDrainsBackendBeforeComposeDown(t *testing.T) {
	logPath := fakeShutdownDocker(t, "present")

	RunStop()

	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	calls := strings.Split(strings.TrimSpace(string(log)), "\n")
	if len(calls) < 3 || calls[1] != "stop --timeout -1 backend-id" || !strings.HasPrefix(calls[2], "compose ") {
		t.Fatalf("backend must drain before compose can stop dependencies: %s", log)
	}
}

func TestRunStopDoesNotStopDependenciesAfterDrainFailure(t *testing.T) {
	for _, mode := range []string{"unavailable", "stop-failed"} {
		t.Run(mode, func(t *testing.T) {
			logPath := fakeShutdownDocker(t, mode)

			RunStop()

			log, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(log), "compose ") || strings.Contains(string(log), "rm ") {
				t.Fatalf("shutdown continued after backend drain failed: %s", log)
			}
		})
	}
}

func TestStopBackendAcceptsAbsentContainer(t *testing.T) {
	logPath := fakeShutdownDocker(t, "absent")

	if err := stopBackend("runtipi"); err != nil {
		t.Fatal(err)
	}

	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(log), "stop ") {
		t.Fatalf("attempted to stop an absent backend: %s", log)
	}
}

func TestStopBackendDockerRecoveryBeyondDefaultTimeout(t *testing.T) {
	image := os.Getenv("RUNTIPI_SHUTDOWN_TEST_IMAGE")
	if image == "" {
		t.Skip("set RUNTIPI_SHUTDOWN_TEST_IMAGE to run the disposable Docker test")
	}

	name := fmt.Sprintf("runtipi-shutdown-test-%d", time.Now().UnixNano())
	process := `trap 'sleep 12; echo recovered; exit 0' TERM; echo ready; while :; do sleep 1; done`
	output, err := exec.Command("docker", "run", "--detach", "--pull", "never", "--name", name, image, "sh", "-c", process).CombinedOutput()
	if err != nil {
		t.Fatalf("create disposable container: %v: %s", err, output)
	}
	t.Cleanup(func() {
		if output, err := exec.Command("docker", "rm", "--force", name).CombinedOutput(); err != nil {
			t.Errorf("remove disposable container: %v: %s", err, output)
		}
	})

	deadline := time.Now().Add(10 * time.Second)
	for {
		output, err = exec.Command("docker", "logs", name).CombinedOutput()
		if err != nil {
			t.Fatalf("read fixture readiness: %v: %s", err, output)
		}
		if strings.Contains(string(output), "ready") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("disposable fixture did not become ready")
		}
		time.Sleep(100 * time.Millisecond)
	}

	started := time.Now()
	if err := stopBackend(name); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(started)

	output, err = exec.Command("docker", "inspect", "--format", "{{.State.ExitCode}}", name).CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "0" {
		t.Fatalf("fixture was killed before recovery: %v: %s", err, output)
	}
	output, err = exec.Command("docker", "logs", name).CombinedOutput()
	if err != nil || !strings.Contains(string(output), "recovered") || elapsed < 12*time.Second {
		t.Fatalf("stop did not wait for recovery (%s): %v: %s", elapsed, err, output)
	}
	t.Logf("backend stop waited %s; recovery marker present and exit code 0", elapsed)
}
