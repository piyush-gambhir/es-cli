package update

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// skippedCommands never trigger the background check: they either report
// versions themselves or produce output other programs consume.
var skippedCommands = map[string]bool{"update": true, "version": true, "completion": true, "help": true}

// NotifierOptions describes the invocation the notifier decides on.
type NotifierOptions struct {
	Version          string
	Commands         []string // names on the invoked command's path, root excluded
	Quiet            bool
	StderrIsTerminal bool
	Getenv           func(string) string
}

// ShouldCheck reports whether this invocation may check for updates. When it
// returns false the notifier makes no network request and prints nothing, so
// scripts, CI, and agents are never disturbed.
func ShouldCheck(o NotifierOptions) bool {
	if !o.StderrIsTerminal || o.Quiet || !IsReleaseVersion(o.Version) {
		return false
	}
	for _, name := range []string{"CI", "ES_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		if o.Getenv(name) != "" {
			return false
		}
	}
	for _, name := range o.Commands {
		if skippedCommands[name] || strings.HasPrefix(name, "__complete") {
			return false
		}
	}
	return true
}

// Check is a background update check. Start it before the command runs and
// call Notify after the command has written its output.
type Check struct {
	current   string
	configDir string
	done      chan struct{}
}

// StartCheck starts an update check. A fresh cached result is used as is;
// otherwise GitHub is queried in a goroutine and the answer (or the failure) is
// cached for 24 hours.
func StartCheck(current, configDir string) *Check {
	c := &Check{current: current, configDir: configDir, done: make(chan struct{})}
	entry := loadCache(configDir)
	if !entry.LastChecked.IsZero() && now().Sub(entry.LastChecked) < cacheTTL {
		close(c.done)
		return c
	}
	go func() {
		defer close(c.done)
		rel, err := FetchLatest(context.Background(), BackgroundTimeout)
		entry := loadCache(configDir)
		entry.LastChecked = now().UTC()
		if err == nil {
			entry.LatestVersion = rel.Version
			entry.PublishedAt = rel.PublishedAt
		}
		saveCache(configDir, entry)
	}()
	return c
}

// Notify prints the update notice to w if the check has already finished, a
// newer release exists, and the notice for that release was not shown in the
// last 24 hours. It never waits for the check.
func (c *Check) Notify(w io.Writer, method InstallMethod) {
	select {
	case <-c.done:
	default:
		return
	}
	entry := loadCache(c.configDir)
	if entry.LatestVersion == "" {
		return
	}
	info := newInfo(c.current, entry.LatestVersion)
	if !info.Available {
		return
	}
	if entry.NotifiedVersion == entry.LatestVersion && now().Sub(entry.NotifiedAt) < cacheTTL {
		return
	}
	PrintNotice(w, info, method)
	entry.NotifiedVersion = entry.LatestVersion
	entry.NotifiedAt = now().UTC()
	saveCache(c.configDir, entry)
}

// PrintNotice writes the new-version notice, preceded by a blank line.
func PrintNotice(w io.Writer, info *Info, method InstallMethod) {
	fmt.Fprintf(w, "\nA new version of %s is available: v%s -> v%s\n", binaryName, info.CurrentVersion, info.LatestVersion)
	fmt.Fprintf(w, "Update with: %s\n", UpdateCommand(method))
	fmt.Fprintf(w, "Release notes: %s\n", ReleaseNotesURL(info.LatestVersion))
}
