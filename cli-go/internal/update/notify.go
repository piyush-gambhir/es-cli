package update

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

var noticeWait = time.Second

var skippedCommands = map[string]bool{"update": true, "version": true, "completion": true, "help": true}

type NotifierOptions struct {
	Version          string
	Commands         []string
	Quiet            bool
	StderrIsTerminal bool
	Getenv           func(string) string
}

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

type Check struct {
	current   string
	configDir string
	done      chan struct{}
}

func StartCheck(current, configDir string) *Check {
	c := &Check{current: current, configDir: configDir, done: make(chan struct{})}
	entry := loadCache(configDir)
	if !entry.LastChecked.IsZero() && now().Sub(entry.LastChecked) < cacheTTL {
		close(c.done)
		return c
	}
	entry.LastChecked = now().UTC()
	saveCache(configDir, entry)
	go func() {
		defer close(c.done)
		rel, err := FetchLatest(context.Background(), BackgroundTimeout)
		if err != nil {
			return
		}
		entry := loadCache(configDir)
		entry.LatestVersion = rel.Version
		saveCache(configDir, entry)
	}()
	return c
}

func (c *Check) Notify(w io.Writer, method InstallMethod) {
	timer := time.NewTimer(noticeWait)
	defer timer.Stop()
	select {
	case <-c.done:
	case <-timer.C:
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

func PrintNotice(w io.Writer, info *Info, method InstallMethod) {
	fmt.Fprintf(w, "\nA new version of %s is available: v%s -> v%s\n", binaryName, info.CurrentVersion, info.LatestVersion)
	fmt.Fprintf(w, "Update with: %s\n", UpdateCommand(method))
	fmt.Fprintf(w, "Release notes: %s\n", ReleaseNotesURL(info.LatestVersion))
}
