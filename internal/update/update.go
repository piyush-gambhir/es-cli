package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	cacheDuration = 24 * time.Hour
	cacheFileName = "update-check.json"
)

// UpdateInfo holds information about an available update.
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	PublishedAt    string
}

// cacheEntry is the on-disk representation of a cached update check.
type cacheEntry struct {
	LastChecked   string `json:"last_checked"`
	LatestVersion string `json:"latest_version"`
	ReleaseURL    string `json:"release_url"`
	PublishedAt   string `json:"published_at,omitempty"`
}

// githubRelease represents the relevant fields from the GitHub releases API.
type githubRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// CheckForUpdate checks GitHub for a newer release. It uses a 24-hour file
// cache to avoid hitting the API on every invocation.
func CheckForUpdate(currentVersion, repo, configDir string) (*UpdateInfo, error) {
	if currentVersion == "" || currentVersion == "dev" {
		return &UpdateInfo{CurrentVersion: currentVersion}, nil
	}

	// Try reading from cache first.
	cachePath := filepath.Join(configDir, cacheFileName)
	if info, err := readCache(cachePath, currentVersion); err == nil && info != nil {
		return info, nil
	}

	// Cache miss or stale — fetch from GitHub.
	return fetchAndCache(currentVersion, repo, cachePath)
}

// CheckForUpdateFresh always checks GitHub, bypassing the cache.
func CheckForUpdateFresh(currentVersion, repo, configDir string) (*UpdateInfo, error) {
	if currentVersion == "" || currentVersion == "dev" {
		return &UpdateInfo{CurrentVersion: currentVersion}, nil
	}

	cachePath := filepath.Join(configDir, cacheFileName)
	return fetchAndCache(currentVersion, repo, cachePath)
}

// PrintUpdateNotice writes an update notice to the given writer.
func PrintUpdateNotice(w io.Writer, info *UpdateInfo) {
	if info == nil || !info.Available {
		return
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "A new version of es is available: %s → %s\n",
		formatVersion(info.CurrentVersion), formatVersion(info.LatestVersion))
	fmt.Fprintf(w, "Run `es update` to update, or download from:\n")
	fmt.Fprintf(w, "%s\n", info.ReleaseURL)
}

func readCache(cachePath, currentVersion string) (*UpdateInfo, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	lastChecked, err := time.Parse(time.RFC3339, entry.LastChecked)
	if err != nil {
		return nil, err
	}

	if time.Since(lastChecked) > cacheDuration {
		return nil, fmt.Errorf("cache expired")
	}

	available, _ := isNewer(entry.LatestVersion, currentVersion)
	return &UpdateInfo{
		Available:      available,
		CurrentVersion: currentVersion,
		LatestVersion:  entry.LatestVersion,
		ReleaseURL:     entry.ReleaseURL,
		PublishedAt:    entry.PublishedAt,
	}, nil
}

func fetchAndCache(currentVersion, repo, cachePath string) (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "es-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding release: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Write cache.
	entry := cacheEntry{
		LastChecked:   time.Now().UTC().Format(time.RFC3339),
		LatestVersion: latestVersion,
		ReleaseURL:    release.HTMLURL,
		PublishedAt:   release.PublishedAt,
	}
	writeCache(cachePath, &entry)

	available, _ := isNewer(latestVersion, currentVersion)
	return &UpdateInfo{
		Available:      available,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseURL:     release.HTMLURL,
		PublishedAt:    release.PublishedAt,
	}, nil
}

func writeCache(cachePath string, entry *cacheEntry) {
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(cachePath, data, 0o644)
}

// isNewer returns true if latest is a higher semver than current.
func isNewer(latest, current string) (bool, error) {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	lMajor, lMinor, lPatch, err := parseSemver(latest)
	if err != nil {
		return false, err
	}
	cMajor, cMinor, cPatch, err := parseSemver(current)
	if err != nil {
		return false, err
	}

	if lMajor != cMajor {
		return lMajor > cMajor, nil
	}
	if lMinor != cMinor {
		return lMinor > cMinor, nil
	}
	return lPatch > cPatch, nil
}

func parseSemver(v string) (major, minor, patch int, err error) {
	// Strip any pre-release or metadata suffix (e.g. "1.2.3-rc1+build").
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver: %s", v)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}
	return major, minor, patch, nil
}

func formatVersion(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}
