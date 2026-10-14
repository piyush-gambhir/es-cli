package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandsResolve(t *testing.T) {
	// Invalid config and a forced terminal ensure the resolver bypasses root setup.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	configDir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "es-cli")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("[invalid yaml"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldTerminal, oldCheck := stderrIsTerminal, startUpdateCheck
	stderrIsTerminal = func() bool { t.Fatal("resolver ran the update notifier hook"); return true }
	t.Cleanup(func() { stderrIsTerminal, startUpdateCheck = oldTerminal, oldCheck })

	for _, tc := range []struct {
		name    string
		args    []string
		want    resolution
		wantErr bool
	}{
		{name: "read", args: []string{"cluster", "health"}, want: resolution{Command: "es cluster health", Effect: "read", Flags: map[string]string{}, Args: []string{}}},
		{name: "write with leading flag", args: []string{"-o", "json", "index", "delete", "my-index", "--confirm"}, want: resolution{Command: "es index delete", Effect: "remote_write", Flags: map[string]string{"output": "json", "confirm": "true"}, Args: []string{"my-index"}}},
		{name: "secrets", args: []string{"config", "view"}, want: resolution{Command: "es config view", Effect: "read", Secret: true, Flags: map[string]string{}, Args: []string{}}},
		{name: "unknown", args: []string{"nonsense"}, wantErr: true},
		{name: "root", wantErr: true},
		{name: "group", args: []string{"index"}, wantErr: true},
		{name: "invalid flag", args: []string{"cluster", "health", "--nonsense"}, wantErr: true},
		{name: "invalid flag value", args: []string{"cluster", "health", "--insecure=invalid"}, wantErr: true},
		{name: "positional separator", args: []string{"document", "get", "my-index", "--", "-id"}, want: resolution{Command: "es document get", Effect: "read", Flags: map[string]string{}, Args: []string{"my-index", "-id"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			root := newRootCmd()
			root.SetOut(&out)
			root.SetErr(&bytes.Buffer{})
			root.SetArgs(append([]string{"commands", "resolve", "--"}, tc.args...))
			err := root.Execute()
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, want error %t", err, tc.wantErr)
			}
			if OutputFormat != "json" {
				t.Fatalf("output format = %q", OutputFormat)
			}
			if tc.wantErr {
				if out.Len() != 0 {
					t.Fatalf("unexpected output: %s", out.String())
				}
				return
			}
			var got resolution
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestResolveCommandAnnotations(t *testing.T) {
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		args        []string
		effect      string
		conditional bool
	}{
		{"read", nil, nil, "read", false},
		{"remote", map[string]string{"mutates": "true", "writes-local": "true"}, nil, "remote_write", false},
		{"conditional remote", map[string]string{"mutates": "conditional"}, nil, "remote_write", true},
		{"remote flag absent", map[string]string{"mutates-when-flag": "write"}, nil, "read", false},
		{"remote flag explicitly false", map[string]string{"mutates-when-flag": "write"}, []string{"--write=false"}, "remote_write", false},
		{"local", map[string]string{"writes-local": "true"}, nil, "local_write", false},
		{"conditional local", map[string]string{"writes-local": "conditional"}, nil, "local_write", true},
		{"self update", map[string]string{"self-update": "true"}, nil, "local_write", false},
		{"local flag absent", map[string]string{"writes-local-when-flag": "write"}, nil, "read", false},
		{"local flag present", map[string]string{"writes-local-when-flag": "write"}, []string{"--write"}, "local_write", false},
		{"nonmutating annotation", map[string]string{"mutates": "false", "writes-local": "false"}, nil, "read", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := &cobra.Command{Use: "probe", Annotations: tc.annotations}
			command.Flags().Bool("write", false, "Write changes")
			if err := command.ParseFlags(tc.args); err != nil {
				t.Fatal(err)
			}
			got := resolveCommand(command)
			if got.Effect != tc.effect || got.Conditional != tc.conditional {
				t.Fatalf("got %+v, want effect %s conditional %t", got, tc.effect, tc.conditional)
			}
		})
	}
}
