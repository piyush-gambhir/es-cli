package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/update"
)

// Test seams: tests replace these to avoid the network, the real executable,
// and the real terminal.
var (
	checkLatest     = update.CheckNow
	executablePath  = update.ExecutablePath
	stdinIsTerminal = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
	installRelease  = func(ctx context.Context, version, exePath string, progress io.Writer) error {
		in := &update.Installer{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, ExePath: exePath, Progress: progress}
		return in.Install(ctx, version)
	}
)

// updateCheckResult is the `es update --check -o json` document.
type updateCheckResult struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url"`
	InstallMethod   string `json:"install_method"`
}

func newUpdateCmd() *cobra.Command {
	var checkOnly, yes bool

	cmd := &cobra.Command{
		Use:         "update",
		Annotations: map[string]string{"mutates": "true"},
		Short:       "Update es to the latest release",
		Long: `Check for and install the latest es release from GitHub Releases.

es update downloads the release archive for this OS and architecture, verifies
its SHA-256 checksum against the release's checksums.txt, and replaces the
running binary. It works on macOS, Linux, and Windows. If the binary's
directory is not writable, it fails and leaves the current binary in place:
re-run with sudo, or reinstall with the install script into a writable directory.
A binary built from source into a Go bin directory is not replaced; es update
prints the command to rebuild it instead.

When stdin is a terminal, es update asks before installing. --yes skips the
prompt; under --no-input, or without a terminal, --yes is required.
--read-only blocks installing; --check is always allowed.

Other commands print a short notice on stderr, at most once a day per release,
when a newer release exists. The notice is only shown in an interactive
terminal and never in CI. Turn it off with ES_NO_UPDATE_NOTIFIER=1 or
NO_UPDATE_NOTIFIER=1.

Examples:
  es update                 # ask, then install the latest release
  es update --yes           # install without asking
  es update --check         # report current and latest versions
  es update --check -o json # same, as JSON`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if checkOnly {
				return runUpdateCheck(cmd)
			}
			return runUpdate(cmd, yes)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only report whether an update is available; do not install")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Install without asking for confirmation")

	return cmd
}

// installMethodOf reports how the binary at exe was installed; an unknown
// path counts as a release install.
func installMethodOf(exe string) update.InstallMethod {
	if exe == "" {
		return update.InstallSelf
	}
	home, _ := os.UserHomeDir()
	return update.DetectInstallMethod(exe, os.Getenv, home)
}

func displayVersion(v string) string {
	if update.IsReleaseVersion(v) {
		return "v" + strings.TrimPrefix(v, "v")
	}
	return v
}

func runUpdateCheck(cmd *cobra.Command) error {
	format := flagOutput
	switch format {
	case "", "table", "json", "yaml":
	default:
		return fmt.Errorf("unsupported output format: %s (use table, json, or yaml)", format)
	}
	OutputFormat = format

	info, err := checkLatest(cmd.Context(), build.Version, config.ConfigDir())
	if err != nil {
		return err
	}
	exe, _ := executablePath()
	method := installMethodOf(exe)
	result := updateCheckResult{
		CurrentVersion:  info.CurrentVersion,
		LatestVersion:   info.LatestVersion,
		UpdateAvailable: info.Available,
		ReleaseURL:      update.ReleaseNotesURL(info.LatestVersion),
		InstallMethod:   string(method),
	}
	if format == "json" || format == "yaml" {
		return output.Print(cmd.OutOrStdout(), format, result, nil)
	}

	w := cmd.OutOrStdout()
	available := "no"
	if info.Available {
		available = "yes"
	}
	fmt.Fprintf(w, "Current version:  %s\n", displayVersion(info.CurrentVersion))
	fmt.Fprintf(w, "Latest version:   v%s\n", info.LatestVersion)
	fmt.Fprintf(w, "Update available: %s\n", available)
	fmt.Fprintf(w, "Release notes:    %s\n", result.ReleaseURL)
	if info.Available {
		fmt.Fprintf(w, "Update with:      %s\n", update.UpdateCommand(method))
	}
	return nil
}

func runUpdate(cmd *cobra.Command, yes bool) error {
	w := cmd.OutOrStdout()
	if !update.IsReleaseVersion(build.Version) {
		fmt.Fprintf(w, "es update is not available for development builds (version %q). Install a release to enable it.\n", build.Version)
		return nil
	}
	if updateBlockedByReadOnly(cmd) {
		return fmt.Errorf("es update is blocked in read-only mode (es update --check still works); remove read_only from the profile, unset ES_READ_ONLY, or drop --read-only to install")
	}

	info, err := checkLatest(cmd.Context(), build.Version, config.ConfigDir())
	if err != nil {
		return err
	}
	notesURL := update.ReleaseNotesURL(info.LatestVersion)
	if !info.Available {
		fmt.Fprintf(w, "es is already up to date (v%s).\n", info.CurrentVersion)
		return nil
	}
	fmt.Fprintf(w, "Update available: v%s -> v%s\n", info.CurrentVersion, info.LatestVersion)

	exe, err := executablePath()
	if err != nil {
		return err
	}
	method := installMethodOf(exe)
	if method == update.InstallGo {
		fmt.Fprintf(w, "es was built from source into %s, so it is not replaced in place.\n", filepath.Dir(exe))
		fmt.Fprintf(w, "Update with: %s\n", update.UpdateCommand(method))
		fmt.Fprintf(w, "Release notes: %s\n", notesURL)
		return nil
	}

	if !yes {
		if flagNoInput || !stdinIsTerminal() {
			return fmt.Errorf("installing es v%s needs confirmation: pass --yes to update without a prompt", info.LatestVersion)
		}
		fmt.Fprint(w, "Update now? [Y/n] ")
		line, readErr := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
		answer := strings.ToLower(strings.TrimSpace(line))
		if (readErr != nil && line == "") || (answer != "" && answer != "y" && answer != "yes") {
			fmt.Fprintln(w, "Update cancelled.")
			return nil
		}
	}

	progress := cmd.ErrOrStderr()
	if flagQuiet {
		progress = io.Discard
	}
	if err := installRelease(cmd.Context(), info.LatestVersion, exe, progress); err != nil {
		return fmt.Errorf("update failed, es v%s was left in place: %w", info.CurrentVersion, err)
	}
	// checkLatest already stored the installed release in the notifier's cache.
	fmt.Fprintf(w, "Updated es v%s -> v%s\n", info.CurrentVersion, info.LatestVersion)
	fmt.Fprintf(w, "Release notes: %s\n", notesURL)
	return nil
}

// updateBlockedByReadOnly applies the same read-only sources as other mutating
// commands: --read-only, ES_READ_ONLY, and the profile's read_only.
func updateBlockedByReadOnly(cmd *cobra.Command) bool {
	if flagReadOnly {
		return true
	}
	if resolved, _, err := loadAndResolveConfig(cmd); err == nil {
		return resolved.ReadOnly
	}
	return envFlagEnabled("ES_READ_ONLY")
}
