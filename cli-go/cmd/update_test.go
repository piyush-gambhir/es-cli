package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/update"
)

type updateStub struct {
	checks     int
	installs   []string // versions passed to installRelease
	installErr error
	exe        string
	configDir  string
}

// stubUpdate runs es as release v0.1.9 with v0.2.0 (or latest) on GitHub,
// without the network, the real executable, or a terminal.
func stubUpdate(t *testing.T, latest string) *updateStub {
	t.Helper()
	s := &updateStub{}
	root := t.TempDir()
	s.exe = filepath.Join(root, "bin", "es")
	if err := os.MkdirAll(filepath.Dir(s.exe), 0o755); err != nil {
		t.Fatal(err)
	}
	xdg := filepath.Join(root, "config")
	s.configDir = filepath.Join(xdg, "es-cli")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("HOME", filepath.Join(root, "home"))
	for _, k := range []string{"GOBIN", "GOPATH", "ES_READ_ONLY", "ES_NO_INPUT", "ES_QUIET", "CI", "ES_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		t.Setenv(k, "")
	}

	origCheck, origExe, origStdinTTY, origStderrTTY, origInstall, origStart := checkLatest, executablePath, stdinIsTerminal, stderrIsTerminal, installRelease, startUpdateCheck
	origVersion, origNoInput, origReadOnly, origOutput, origQuiet := build.Version, flagNoInput, flagReadOnly, flagOutput, flagQuiet
	t.Cleanup(func() {
		checkLatest, executablePath, stdinIsTerminal, stderrIsTerminal, installRelease, startUpdateCheck = origCheck, origExe, origStdinTTY, origStderrTTY, origInstall, origStart
		build.Version, flagNoInput, flagReadOnly, flagOutput, flagQuiet = origVersion, origNoInput, origReadOnly, origOutput, origQuiet
	})

	build.Version = "0.1.9"
	flagNoInput, flagReadOnly, flagOutput, flagQuiet = false, false, "", false
	checkLatest = func(_ context.Context, current, _ string) (*update.Info, error) {
		s.checks++
		return &update.Info{CurrentVersion: current, LatestVersion: latest, Available: latest != current}, nil
	}
	executablePath = func() (string, error) { return s.exe, nil }
	stdinIsTerminal = func() bool { return false }
	stderrIsTerminal = func() bool { return false }
	startUpdateCheck = func(string, string) *update.Check { t.Fatal("unexpected background update check"); return nil }
	installRelease = func(_ context.Context, version, exePath string, _ io.Writer) error {
		if exePath != s.exe {
			t.Fatalf("installing over %q, want %q", exePath, s.exe)
		}
		s.installs = append(s.installs, version)
		return s.installErr
	}
	return s
}

func runES(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestUpdateCheckJSON(t *testing.T) {
	stubUpdate(t, "0.2.0")
	out, err := runES(t, "", "update", "--check", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	want := map[string]any{
		"current_version":  "0.1.9",
		"latest_version":   "0.2.0",
		"update_available": true,
		"release_url":      "https://github.com/piyush-gambhir/es-cli/releases/tag/v0.2.0",
		"install_method":   "self",
	}
	if len(got) != len(want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %v, want %v", k, got[k], v)
		}
	}
}

func TestUpdateCheckReportsGoInstall(t *testing.T) {
	s := stubUpdate(t, "0.2.0")
	t.Setenv("GOBIN", filepath.Dir(s.exe))
	out, err := runES(t, "", "update", "--check", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"install_method": "go"`) {
		t.Fatalf("install_method is not go:\n%s", out)
	}
}

func TestUpdateCheckText(t *testing.T) {
	stubUpdate(t, "0.1.9")
	out, err := runES(t, "", "update", "--check", "--read-only")
	if err != nil {
		t.Fatalf("update --check under --read-only: %v", err)
	}
	for _, want := range []string{"Current version:  v0.1.9", "Latest version:   v0.1.9", "Update available: no", "releases/tag/v0.1.9"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestUpdateAlreadyLatest(t *testing.T) {
	s := stubUpdate(t, "0.1.9")
	out, err := runES(t, "", "update")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "already up to date (v0.1.9)") || len(s.installs) != 0 {
		t.Fatalf("installs=%v output:\n%s", s.installs, out)
	}
}

func TestUpdateYesInstallsAndKeepsCache(t *testing.T) {
	s := stubUpdate(t, "0.2.0")
	if err := os.MkdirAll(s.configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(s.configDir, "update-check.json")
	if err := os.WriteFile(cache, []byte(`{"latest_version":"0.2.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := runES(t, "", "update", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.installs) != 1 || s.installs[0] != "0.2.0" {
		t.Fatalf("installs = %v, want [0.2.0]", s.installs)
	}
	for _, want := range []string{"Updated es v0.1.9 -> v0.2.0\n", "Release notes: https://github.com/piyush-gambhir/es-cli/releases/tag/v0.2.0\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	// The cache holds the installed release, so the notifier stays quiet and
	// es version agrees with es update.
	if data, err := os.ReadFile(cache); err != nil || !strings.Contains(string(data), `"latest_version":"0.2.0"`) {
		t.Fatalf("update cache after a successful update = %q, %v", data, err)
	}
}

func TestUpdateNeedsYesWithoutPrompt(t *testing.T) {
	for name, args := range map[string][]string{
		"--no-input":    {"update", "--no-input"},
		"stdin not tty": {"update"},
	} {
		t.Run(name, func(t *testing.T) {
			s := stubUpdate(t, "0.2.0")
			_, err := runES(t, "y\n", args...)
			if err == nil || !strings.Contains(err.Error(), "--yes") {
				t.Fatalf("error = %v, want one that says to pass --yes", err)
			}
			if len(s.installs) != 0 {
				t.Fatalf("installed without confirmation: %v", s.installs)
			}
		})
	}
}

func TestUpdatePrompt(t *testing.T) {
	for stdin, wantInstall := range map[string]bool{"\n": true, "y\n": true, "YES\n": true, "n\n": false, "": false} {
		s := stubUpdate(t, "0.2.0")
		stdinIsTerminal = func() bool { return true }
		out, err := runES(t, stdin, "update")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "Update available: v0.1.9 -> v0.2.0\nUpdate now? [Y/n] ") {
			t.Fatalf("stdin %q: no prompt:\n%s", stdin, out)
		}
		if got := len(s.installs) == 1; got != wantInstall {
			t.Errorf("stdin %q: installed = %t, want %t", stdin, got, wantInstall)
		}
	}
}

func TestUpdateReadOnlyBlocksInstall(t *testing.T) {
	for name, set := range map[string]func(){
		"flag": func() {},
		"env":  func() { t.Setenv("ES_READ_ONLY", "true") },
	} {
		t.Run(name, func(t *testing.T) {
			s := stubUpdate(t, "0.2.0")
			set()
			args := []string{"update", "--yes"}
			if name == "flag" {
				args = append(args, "--read-only")
			}
			_, err := runES(t, "", args...)
			if err == nil || !strings.Contains(err.Error(), "read-only") {
				t.Fatalf("error = %v, want a read-only refusal", err)
			}
			if s.checks != 0 || len(s.installs) != 0 {
				t.Fatalf("checks=%d installs=%v under read-only", s.checks, s.installs)
			}
		})
	}
}

func TestUpdateGoInstallPrintsCommand(t *testing.T) {
	s := stubUpdate(t, "0.2.0")
	t.Setenv("GOPATH", filepath.Dir(filepath.Dir(s.exe)))
	out, err := runES(t, "", "update", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Update with: git pull && make install") || len(s.installs) != 0 {
		t.Fatalf("installs=%v output:\n%s", s.installs, out)
	}
}

func TestUpdateFailureKeepsOldVersion(t *testing.T) {
	s := stubUpdate(t, "0.2.0")
	s.installErr = errors.New("SHA-256 mismatch")
	_, err := runES(t, "", "update", "--yes")
	if err == nil || !strings.Contains(err.Error(), "es v0.1.9 was left in place") || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("error = %v", err)
	}
}

func TestVersionShowsCachedLatestOnly(t *testing.T) {
	s := stubUpdate(t, "0.2.0")
	checkLatest = func(context.Context, string, string) (*update.Info, error) {
		t.Fatal("version contacted GitHub")
		return nil, nil
	}

	out, err := runES(t, "", "version")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "latest") {
		t.Fatalf("version printed latest without a cache:\n%s", out)
	}

	if err := os.MkdirAll(s.configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.configDir, "update-check.json"), []byte(`{"latest_version":"0.2.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err = runES(t, "", "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "es-cli version 0.1.9\n") || !strings.Contains(out, "  latest: 0.2.0\n  update_available: true\n") {
		t.Fatalf("version output:\n%s", out)
	}
}

// The notifier must not even start (no network, no output) unless the
// invocation is an interactive release build with no opt-out.
func TestNotifierSuppression(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		setup func(t *testing.T)
		want  bool
	}{
		{"interactive", []string{"config", "list-profiles"}, nil, true},
		{"stderr not a terminal", []string{"config", "list-profiles"}, func(*testing.T) { stderrIsTerminal = func() bool { return false } }, false},
		{"CI", []string{"config", "list-profiles"}, func(t *testing.T) { t.Setenv("CI", "true") }, false},
		{"ES_NO_UPDATE_NOTIFIER", []string{"config", "list-profiles"}, func(t *testing.T) { t.Setenv("ES_NO_UPDATE_NOTIFIER", "1") }, false},
		{"NO_UPDATE_NOTIFIER", []string{"config", "list-profiles"}, func(t *testing.T) { t.Setenv("NO_UPDATE_NOTIFIER", "1") }, false},
		{"--quiet", []string{"config", "list-profiles", "--quiet"}, nil, false},
		{"ES_QUIET", []string{"config", "list-profiles"}, func(t *testing.T) { t.Setenv("ES_QUIET", "1") }, false},
		{"dev build", []string{"config", "list-profiles"}, func(*testing.T) { build.Version = "dev" }, false},
		{"version", []string{"version"}, nil, false},
		{"update", []string{"update", "--check"}, nil, false},
		{"completion", []string{"completion", "bash"}, discardStdout, false},
		{"help", []string{"help"}, nil, false},
		{"__complete", []string{"__complete", "config", ""}, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubUpdate(t, "0.2.0")
			stderrIsTerminal = func() bool { return true }
			started := false
			startUpdateCheck = func(string, string) *update.Check { started = true; return nil }
			if tc.setup != nil {
				tc.setup(t)
			}
			if _, err := runES(t, "", tc.args...); err != nil {
				t.Fatal(err)
			}
			if started != tc.want {
				t.Fatalf("background check started = %t, want %t", started, tc.want)
			}
		})
	}
}

// discardStdout hides output that commands write straight to os.Stdout.
func discardStdout(t *testing.T) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = devNull
	t.Cleanup(func() { os.Stdout = orig; devNull.Close() })
}
