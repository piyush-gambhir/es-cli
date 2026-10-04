package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/update"
)

// A traversing entry name must not place the binary outside destDir.
func TestExtractBinaryIgnoresEntryPath(t *testing.T) {
	root := t.TempDir()
	destDir := filepath.Join(root, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(root, "es-cli.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	body := []byte("new binary")
	if err := tw.WriteHeader(&tar.Header{Name: "../../es-cli", Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	for _, c := range []interface{ Close() error }{tw, gz, f} {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}

	got, err := extractBinary(archivePath, destDir)
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if want := filepath.Join(destDir, "es"); got != want {
		t.Fatalf("extracted to %q, want %q", got, want)
	}
	if data, err := os.ReadFile(got); err != nil || string(data) != string(body) {
		t.Fatalf("extracted contents = %q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(root, "es-cli")); !os.IsNotExist(err) {
		t.Fatalf("entry path escaped destDir: %v", err)
	}
}

// stubWindowsUpdate makes `es update` see a newer release while running as a
// Windows release build, without touching the network or the real stdin.
func stubWindowsUpdate(t *testing.T, stdin string) string {
	t.Helper()
	const releaseURL = "https://github.com/piyush-gambhir/es-cli/releases/tag/v9.9.9"
	origGOOS, origCheck, origVersion, origStdin, origNoInput := goos, checkForUpdate, build.Version, os.Stdin, flagNoInput
	t.Cleanup(func() {
		goos, checkForUpdate, build.Version, os.Stdin, flagNoInput = origGOOS, origCheck, origVersion, origStdin, origNoInput
	})

	goos = "windows"
	build.Version = "0.1.0"
	flagNoInput = false
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	checkForUpdate = func(current, repo, configDir string) (*update.UpdateInfo, error) {
		return &update.UpdateInfo{Available: true, CurrentVersion: current, LatestVersion: "9.9.9", ReleaseURL: releaseURL}, nil
	}

	in, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.WriteString(stdin); err != nil {
		t.Fatal(err)
	}
	if _, err := in.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { in.Close() })
	os.Stdin = in
	return releaseURL
}

// Windows releases ship as .zip and a running .exe cannot be replaced, so the
// install path must refuse (before prompting) and point at the release page.
func TestUpdateRefusesInstallOnWindows(t *testing.T) {
	releaseURL := stubWindowsUpdate(t, "n\n")

	cmd := newUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("update on Windows succeeded, want an error; output:\n%s", out.String())
	}
	if !strings.Contains(err.Error(), "es.exe") || !strings.Contains(err.Error(), releaseURL) {
		t.Fatalf("error = %q, want it to name es.exe and %s", err, releaseURL)
	}
	if strings.Contains(out.String(), "Do you want to update?") {
		t.Fatalf("prompted before refusing; output:\n%s", out.String())
	}
}

func TestUpdateCheckWorksOnWindows(t *testing.T) {
	releaseURL := stubWindowsUpdate(t, "")

	cmd := newUpdateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update --check on Windows: %v", err)
	}
	if !strings.Contains(out.String(), "v9.9.9") || !strings.Contains(out.String(), releaseURL) {
		t.Fatalf("update --check output missing the new version or release URL:\n%s", out.String())
	}
}
