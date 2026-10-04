package update

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClock pins now() for the test and returns a function that moves it.
func fakeClock(t *testing.T) func(time.Duration) {
	t.Helper()
	current := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	orig := now
	now = func() time.Time { return current }
	t.Cleanup(func() { now = orig })
	return func(d time.Duration) { current = current.Add(d) }
}

// latestServer serves GitHub's releases/latest endpoint with *tag and counts
// requests. An empty tag answers HTTP 500.
func latestServer(t *testing.T, tag *string) *atomic.Int32 {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/"+Repo+"/releases/latest" {
			http.NotFound(w, r)
			return
		}
		hits.Add(1)
		if *tag == "" {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, `{"tag_name":%q,"published_at":"2026-10-01T00:00:00Z"}`, *tag)
	}))
	t.Cleanup(srv.Close)
	orig := apiBaseURL
	apiBaseURL = srv.URL
	t.Cleanup(func() { apiBaseURL = orig })
	return &hits
}

func runCheck(t *testing.T, current, dir string, method InstallMethod) string {
	t.Helper()
	c := StartCheck(current, dir)
	select {
	case <-c.done:
	case <-time.After(5 * time.Second):
		t.Fatal("update check did not finish")
	}
	var buf bytes.Buffer
	c.Notify(&buf, method)
	return buf.String()
}

func TestShouldCheckSuppression(t *testing.T) {
	env := map[string]string{}
	base := func() NotifierOptions {
		return NotifierOptions{
			Version:          "0.1.9",
			Commands:         []string{"list", "index"},
			StderrIsTerminal: true,
			Getenv:           func(k string) string { return env[k] },
		}
	}
	if !ShouldCheck(base()) {
		t.Fatal("interactive release build with no opt-out should check")
	}

	for name, mutate := range map[string]func(*NotifierOptions){
		"stderr not a terminal": func(o *NotifierOptions) { o.StderrIsTerminal = false },
		"quiet":                 func(o *NotifierOptions) { o.Quiet = true },
		"dev build":             func(o *NotifierOptions) { o.Version = "dev" },
		"empty version":         func(o *NotifierOptions) { o.Version = "" },
		"commit-hash version":   func(o *NotifierOptions) { o.Version = "7433542" },
		"update command":        func(o *NotifierOptions) { o.Commands = []string{"update"} },
		"version command":       func(o *NotifierOptions) { o.Commands = []string{"version"} },
		"completion subcommand": func(o *NotifierOptions) { o.Commands = []string{"bash", "completion"} },
		"help command":          func(o *NotifierOptions) { o.Commands = []string{"help"} },
		"__complete":            func(o *NotifierOptions) { o.Commands = []string{"__complete"} },
		"__completeNoDesc":      func(o *NotifierOptions) { o.Commands = []string{"__completeNoDesc"} },
	} {
		o := base()
		mutate(&o)
		if ShouldCheck(o) {
			t.Errorf("%s: ShouldCheck = true, want false", name)
		}
	}

	for _, key := range []string{"CI", "ES_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		for _, value := range []string{"1", "true", "0"} {
			env = map[string]string{key: value}
			if ShouldCheck(base()) {
				t.Errorf("%s=%s: ShouldCheck = true, want false", key, value)
			}
		}
	}
}

func TestNoticeShownOncePerVersionPerDay(t *testing.T) {
	advance := fakeClock(t)
	tag := "v0.2.0"
	hits := latestServer(t, &tag)
	dir := t.TempDir()

	want := "\nA new version of es is available: v0.1.9 -> v0.2.0\n" +
		"Update with: es update\n" +
		"Release notes: https://github.com/piyush-gambhir/es-cli/releases/tag/v0.2.0\n"
	if got := runCheck(t, "0.1.9", dir, InstallSelf); got != want {
		t.Fatalf("first notice = %q, want %q", got, want)
	}

	advance(time.Hour)
	if got := runCheck(t, "0.1.9", dir, InstallSelf); got != "" {
		t.Fatalf("same version within 24h printed again: %q", got)
	}
	if n := hits.Load(); n != 1 {
		t.Fatalf("GitHub hit %d times within the cache window, want 1", n)
	}

	// A newer release is announced even inside the 24h window.
	entry := loadCache(dir)
	entry.LatestVersion = "0.2.1"
	saveCache(dir, entry)
	if got := runCheck(t, "0.1.9", dir, InstallSelf); !strings.Contains(got, "v0.1.9 -> v0.2.1") {
		t.Fatalf("new version within 24h not announced: %q", got)
	}

	// After 24h the cache expires, GitHub is asked again, and the notice repeats.
	advance(25 * time.Hour)
	if got := runCheck(t, "0.1.9", dir, InstallSelf); !strings.Contains(got, "v0.1.9 -> v0.2.0") {
		t.Fatalf("notice not repeated after 24h: %q", got)
	}
	if n := hits.Load(); n != 2 {
		t.Fatalf("GitHub hit %d times, want 2 after the cache expired", n)
	}
}

func TestNoticeUsesSourceBuildCommandForGoInstalls(t *testing.T) {
	var buf bytes.Buffer
	PrintNotice(&buf, &Info{CurrentVersion: "0.1.9", LatestVersion: "0.2.0", Available: true}, InstallGo)
	if !strings.Contains(buf.String(), "\nUpdate with: git pull && make install (in your es-cli/cli-go checkout)\n") {
		t.Fatalf("go-install notice has the wrong update line:\n%s", buf.String())
	}
	if strings.Contains(buf.String(), "es update") {
		t.Fatalf("go-install notice suggests es update:\n%s", buf.String())
	}
}

func TestNoNoticeWhenUpToDate(t *testing.T) {
	fakeClock(t)
	tag := "v0.1.9"
	latestServer(t, &tag)
	if got := runCheck(t, "0.1.9", t.TempDir(), InstallSelf); got != "" {
		t.Fatalf("notice for the running version: %q", got)
	}
}

func TestFailedCheckIsCached(t *testing.T) {
	advance := fakeClock(t)
	tag := ""
	hits := latestServer(t, &tag)
	dir := t.TempDir()

	if got := runCheck(t, "0.1.9", dir, InstallSelf); got != "" {
		t.Fatalf("failed check printed %q", got)
	}
	advance(time.Hour)
	runCheck(t, "0.1.9", dir, InstallSelf)
	if n := hits.Load(); n != 1 {
		t.Fatalf("failed check retried: %d requests, want 1", n)
	}
}

// Notify must not wait for a check that is still running.
func TestNotifyNeverWaits(t *testing.T) {
	fakeClock(t)
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
		fmt.Fprint(w, `{"tag_name":"v0.2.0"}`)
	}))
	t.Cleanup(srv.Close)
	orig := apiBaseURL
	apiBaseURL = srv.URL
	t.Cleanup(func() { apiBaseURL = orig })

	c := StartCheck("0.1.9", t.TempDir())
	var buf bytes.Buffer
	start := time.Now()
	c.Notify(&buf, InstallSelf)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("Notify blocked for %s", elapsed)
	}
	if buf.Len() != 0 {
		t.Fatalf("Notify printed before the check finished: %q", buf.String())
	}
	close(release)
	<-c.done
}

func TestCheckNowBypassesCache(t *testing.T) {
	fakeClock(t)
	tag := "v0.2.0"
	hits := latestServer(t, &tag)
	dir := t.TempDir()
	saveCache(dir, cacheEntry{LastChecked: now(), LatestVersion: "0.1.9"})

	info, err := CheckNow(context.Background(), "0.1.9", dir)
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 || info.LatestVersion != "0.2.0" || !info.Available {
		t.Fatalf("CheckNow = %+v after %d requests, want 0.2.0 from GitHub", info, hits.Load())
	}
	if cached, ok := Cached("0.1.9", dir); !ok || cached.LatestVersion != "0.2.0" {
		t.Fatalf("cache after CheckNow = %+v, %t", cached, ok)
	}
}

func TestFetchLatestRejectsUnexpectedTag(t *testing.T) {
	tag := "v1.0.0/../../evil"
	latestServer(t, &tag)
	if _, err := FetchLatest(context.Background(), time.Second); err == nil {
		t.Fatal("FetchLatest accepted a tag that is not a version")
	}
}

func TestDetectInstallMethod(t *testing.T) {
	root := t.TempDir()
	dirs := map[string]string{}
	for _, name := range []string{"gobin", "gopath1/bin", "gopath2/bin", "home/go/bin", "usr/local/bin", "home/Go/bin"} {
		dirs[name] = filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(dirs[name], 0o755); err != nil {
			t.Fatal(err)
		}
	}
	env := map[string]string{
		"GOBIN":  dirs["gobin"],
		"GOPATH": filepath.Join(root, "gopath1") + string(os.PathListSeparator) + filepath.Join(root, "gopath2"),
	}
	getenv := func(k string) string { return env[k] }
	home := filepath.Join(root, "home")

	for dir, want := range map[string]InstallMethod{
		"gobin":         InstallGo,
		"gopath1/bin":   InstallGo,
		"gopath2/bin":   InstallGo,
		"home/go/bin":   InstallGo,
		"usr/local/bin": InstallSelf,
	} {
		if got := DetectInstallMethod(filepath.Join(dirs[dir], "es"), getenv, home); got != want {
			t.Errorf("binary in %s: install method %q, want %q", dir, got, want)
		}
	}

	// On a case-sensitive filesystem home/Go/bin is a different directory from
	// home/go/bin and must not count as a Go bin directory.
	goBin, _ := os.Stat(dirs["home/go/bin"])
	upper, _ := os.Stat(dirs["home/Go/bin"])
	if !os.SameFile(goBin, upper) {
		if got := DetectInstallMethod(filepath.Join(dirs["home/Go/bin"], "es"), getenv, home); got != InstallSelf {
			t.Errorf("binary in home/Go/bin: install method %q, want self", got)
		}
	}
}
