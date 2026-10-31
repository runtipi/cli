package utils_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/runtipi/cli/internal/config"
	"github.com/runtipi/cli/internal/utils"
)

func installation(t *testing.T) string {
	t.Helper()
	previousRoot := config.RootFolder
	config.RootFolder = t.TempDir()
	t.Cleanup(func() { config.RootFolder = previousRoot })
	return config.RootFolder
}

func dockerCommand(t *testing.T, script string) {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "docker"), []byte("#!/bin/sh\n"+script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func assertMode(t *testing.T, path string, expected os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != expected {
		t.Fatalf("%s mode = %o, want %o", path, info.Mode().Perm(), expected)
	}
}

func TestSharedPermissionRepairPreservesAccessAndSecrets(t *testing.T) {
	root := installation(t)
	if err := os.Chmod(filepath.Dir(root), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0755); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{"state", "apps/demo", "repos", "logs", "user-config", "traefik/shared", "data/postgres", "app-data/demo"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range map[string]string{".env": "JWT_SECRET=existing-encryption-key\nPOSTGRES_PASSWORD=preserved\n", "state/seed": "existing-seed", "state/settings.json": "{}", "apps/demo/config": "host-readable", "user-config/tipi-compose.yml": "host-compose", "traefik/shared/acme.json": "{}"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := utils.CopySystemFiles(); err != nil {
		t.Fatal(err)
	}
	if err := utils.GenerateEnvFile(""); err != nil {
		t.Fatal(err)
	}
	private := t.TempDir()
	if err := os.Chmod(private, 0700); err != nil {
		t.Fatal(err)
	}
	key := filepath.Join(private, "cli-secret")
	if err := os.WriteFile(key, []byte(strings.Repeat("c", 64)), 0600); err != nil {
		t.Fatal(err)
	}

	assertOtherUIDCannotRead := func() {
		if runtime.GOOS != "linux" || os.Geteuid() != 0 {
			return
		}
		command := exec.Command("cat", key)
		command.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: 65534, Gid: 65534}}
		if err := command.Run(); err == nil {
			t.Fatal("another UID read the CLI key")
		}
	}
	assertOtherUIDCannotRead()
	if err := utils.EnsureFilePermissions(); err != nil {
		t.Fatal(err)
	}
	assertOtherUIDCannotRead()
	for _, name := range []string{"state", "apps/demo/config", "user-config/tipi-compose.yml", "repos", "logs", "traefik", "data/postgres"} {
		assertMode(t, filepath.Join(root, name), 0777)
	}
	assertMode(t, root, 0755)
	assertMode(t, private, 0700)
	assertMode(t, key, 0600)
	if runtime.GOOS == "linux" && os.Geteuid() == 0 {
		command := exec.Command("cat", filepath.Join(root, "apps/demo/config"), filepath.Join(root, "user-config/tipi-compose.yml"))
		command.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: 65534, Gid: 65534}}
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("host user lost shared file access: %s %v", output, err)
		}
	}
	if utils.GetEnvValue("JWT_SECRET") != "existing-encryption-key" || utils.GetEnvValue("POSTGRES_PASSWORD") != "preserved" {
		t.Fatal("existing credentials changed")
	}
	content, err := os.ReadFile(filepath.Join(root, "state/seed"))
	if err != nil || string(content) != "existing-seed" {
		t.Fatalf("seed changed: %v", err)
	}
}

func TestCliTokenUsesDockerKey(t *testing.T) {
	installation(t)
	secret := strings.Repeat("c", 64)
	dockerCommand(t, `test "$*" = "exec runtipi cat /data/.secrets/cli-secret" || exit 2
printf '%s' '`+secret+`'
`)

	tokenString, err := utils.CreateToken()
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) { return []byte(secret), nil }, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		t.Fatalf("private key did not verify token: %v", err)
	}
	if subject, err := token.Claims.GetSubject(); err != nil || subject != "cli" {
		t.Fatalf("unexpected subject: %q %v", subject, err)
	}
	if _, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) { return []byte("legacy-secret"), nil }); err == nil {
		t.Fatal("legacy key verified CLI token")
	}
}

func TestCliTokenRejectsUnavailableOrMalformedDockerKey(t *testing.T) {
	for _, output := range []string{"", "legacy-secret", strings.Repeat("a", 63), strings.Repeat("g", 64), strings.Repeat("A", 64), strings.Repeat("a", 64) + "\n"} {
		t.Run(output, func(t *testing.T) {
			dockerCommand(t, "printf '%s' '"+output+"'\n")
			if _, err := utils.CreateToken(); err == nil || !strings.Contains(err.Error(), "invalid CLI authentication key") {
				t.Fatalf("malformed Docker key accepted: %v", err)
			}
		})
	}

	t.Run("unavailable", func(t *testing.T) {
		dockerCommand(t, "exit 1\n")
		if _, err := utils.CreateToken(); err == nil || !strings.Contains(err.Error(), "start the updated backend") {
			t.Fatalf("missing backend error = %v", err)
		}
	})
}

func TestAPIRequestCancellationIncludesDockerRead(t *testing.T) {
	dockerCommand(t, "exec sleep 30\n")
	for _, canceled := range []bool{false, true} {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		if canceled {
			cancel()
		}
		started := time.Now()
		_, err := utils.APIRequestWithContext(ctx, "http://127.0.0.1:1", "GET", "")
		cancel()
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("cancellation error = %v", err)
		}
		if time.Since(started) > 2*time.Second {
			t.Fatal("Docker key read ignored request cancellation")
		}
	}
}

func TestAPIResponseBodyRemainsReadableAfterDockerAuthentication(t *testing.T) {
	dockerCommand(t, "printf '%s' '"+strings.Repeat("c", 64)+"'\n")
	finish := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.(http.Flusher).Flush()
		<-finish
		_, _ = writer.Write([]byte("authenticated"))
	}))
	defer server.Close()

	response, err := utils.APIRequestWithContext(context.Background(), server.URL, "GET", "")
	close(finish)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil || string(body) != "authenticated" {
		t.Fatalf("response body = %q, error = %v", body, err)
	}
}

func TestNonRootDockerCallerDoesNotNeedHostKeyAccess(t *testing.T) {
	if os.Getenv("RUNTIPI_NONROOT_AUTH_TEST") == "1" {
		if _, err := os.ReadFile(os.Getenv("RUNTIPI_PRIVATE_HOST_KEY")); err == nil {
			t.Fatal("non-root caller read root-owned host key")
		}
		if _, err := utils.CreateToken(); err != nil {
			t.Fatal(err)
		}
		return
	}
	if runtime.GOOS != "linux" || os.Geteuid() != 0 {
		t.Skip("requires Linux root to exercise a separate host UID")
	}

	fixture := t.TempDir()
	for _, path := range []string{filepath.Dir(fixture), fixture} {
		if err := os.Chmod(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	key := filepath.Join(fixture, "root-owned-key")
	if err := os.WriteFile(key, []byte(strings.Repeat("c", 64)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture, "docker"), []byte("#!/bin/sh\nprintf '%s' '"+strings.Repeat("c", 64)+"'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(fixture, "auth.test")
	if err := os.WriteFile(child, binary, 0755); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(child, "-test.run=^TestNonRootDockerCallerDoesNotNeedHostKeyAccess$")
	command.Env = append(os.Environ(), "RUNTIPI_NONROOT_AUTH_TEST=1", "RUNTIPI_PRIVATE_HOST_KEY="+key, "PATH="+fixture+":"+os.Getenv("PATH"))
	command.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: 1000, Gid: 1000}}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("non-root Docker caller failed: %s %v", output, err)
	}
}
