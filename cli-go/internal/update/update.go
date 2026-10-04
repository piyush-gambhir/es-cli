// Package update looks up the latest es release on GitHub, caches the result,
// prints the new-version notice, and installs releases in place.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// Repo is the GitHub repository that publishes es releases.
	Repo = "piyush-gambhir/es-cli"

	binaryName    = "es"
	projectName   = "es-cli"
	cacheFileName = "update-check.json"
	cacheTTL      = 24 * time.Hour

	// BackgroundTimeout bounds the notifier's GitHub request.
	BackgroundTimeout = 3 * time.Second
	// ForegroundTimeout bounds the GitHub request made by `es update`.
	ForegroundTimeout = 15 * time.Second
)

// Test seams: tests point these at httptest servers and fix the clock.
var (
	apiBaseURL      = "https://api.github.com"
	downloadBaseURL = "https://github.com"
	now             = time.Now
)

// Release is the latest published release.
type Release struct {
	Version     string // without the "v" prefix
	PublishedAt string
}

// Info compares the running version with the latest known release.
type Info struct {
	CurrentVersion string
	LatestVersion  string
	Available      bool
}

// cacheEntry is update-check.json in the config dir. A failed check still
// records LastChecked so a broken network does not cause a request per command.
type cacheEntry struct {
	LastChecked     time.Time `json:"last_checked,omitzero"`
	LatestVersion   string    `json:"latest_version,omitempty"`
	PublishedAt     string    `json:"published_at,omitempty"`
	NotifiedVersion string    `json:"notified_version,omitempty"`
	NotifiedAt      time.Time `json:"notified_at,omitzero"`
}

var semverPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// IsReleaseVersion reports whether v is a semantic version, as stamped into
// release builds. "dev", empty, and bare commit hashes are not.
func IsReleaseVersion(v string) bool {
	return semverPattern.MatchString(v)
}

// ReleaseNotesURL is the GitHub release page for version.
func ReleaseNotesURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", Repo, strings.TrimPrefix(version, "v"))
}

// FetchLatest asks GitHub for the latest release.
func FetchLatest(ctx context.Context, timeout time.Duration) (*Release, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/repos/"+Repo+"/releases/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", projectName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checking for updates: GitHub returned HTTP %d", resp.StatusCode)
	}

	var body struct {
		TagName     string `json:"tag_name"`
		PublishedAt string `json:"published_at"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding release: %w", err)
	}
	// The tag ends up in download URLs, so accept only a plain version.
	if !IsReleaseVersion(body.TagName) {
		return nil, fmt.Errorf("latest release has an unexpected tag %q", body.TagName)
	}
	return &Release{Version: strings.TrimPrefix(body.TagName, "v"), PublishedAt: body.PublishedAt}, nil
}

// CheckNow queries GitHub (ignoring any cached result), records the answer in
// the cache, and compares it with current.
func CheckNow(ctx context.Context, current, configDir string) (*Info, error) {
	rel, err := FetchLatest(ctx, ForegroundTimeout)
	if err != nil {
		return nil, err
	}
	entry := loadCache(configDir)
	entry.LastChecked = now().UTC()
	entry.LatestVersion = rel.Version
	entry.PublishedAt = rel.PublishedAt
	saveCache(configDir, entry)
	return newInfo(current, rel.Version), nil
}

// Cached returns what the cache knows about the latest release, without any
// network access. ok is false when nothing is cached.
func Cached(current, configDir string) (info *Info, ok bool) {
	entry := loadCache(configDir)
	if entry.LatestVersion == "" {
		return nil, false
	}
	return newInfo(current, entry.LatestVersion), true
}

// ClearCache removes the cached check, for example after an update.
func ClearCache(configDir string) {
	_ = os.Remove(filepath.Join(configDir, cacheFileName))
}

func newInfo(current, latest string) *Info {
	available, _ := isNewer(latest, current)
	return &Info{CurrentVersion: strings.TrimPrefix(current, "v"), LatestVersion: latest, Available: available}
}

func loadCache(configDir string) cacheEntry {
	var entry cacheEntry
	data, err := os.ReadFile(filepath.Join(configDir, cacheFileName))
	if err != nil {
		return cacheEntry{}
	}
	if json.Unmarshal(data, &entry) != nil {
		return cacheEntry{}
	}
	return entry
}

// saveCache writes the cache through a temp file and rename, so a process that
// exits mid-write never leaves a truncated file behind. Errors are ignored: the
// cache is an optimization.
func saveCache(configDir string, entry cacheEntry) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(configDir, ".update-check-*")
	if err != nil {
		return
	}
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil || cerr != nil || os.Rename(tmp.Name(), filepath.Join(configDir, cacheFileName)) != nil {
		_ = os.Remove(tmp.Name())
	}
}

// isNewer reports whether latest is a higher version than current. Pre-release
// and build suffixes are ignored, so a source build described as
// v0.1.9-3-gabc123 counts as 0.1.9.
func isNewer(latest, current string) (bool, error) {
	l, err := parseSemver(latest)
	if err != nil {
		return false, err
	}
	c, err := parseSemver(current)
	if err != nil {
		return false, err
	}
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i], nil
		}
	}
	return false, nil
}

func parseSemver(v string) ([3]int, error) {
	var out [3]int
	v = strings.TrimPrefix(v, "v")
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, fmt.Errorf("invalid version %q", v)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out, fmt.Errorf("invalid version %q", v)
		}
		out[i] = n
	}
	return out, nil
}
