package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	Repo = "piyush-gambhir/es-cli"

	binaryName    = "es"
	projectName   = "es-cli"
	cacheFileName = "update-check.json"
	cacheTTL      = 24 * time.Hour

	BackgroundTimeout = 3 * time.Second
	ForegroundTimeout = 15 * time.Second
)

var (
	downloadBaseURL = "https://github.com"
	now             = time.Now
)

type Release struct {
	Version string
}

type Info struct {
	CurrentVersion string
	LatestVersion  string
	Available      bool
}

type cacheEntry struct {
	LastChecked     time.Time `json:"last_checked,omitzero"`
	LatestVersion   string    `json:"latest_version,omitempty"`
	NotifiedVersion string    `json:"notified_version,omitempty"`
	NotifiedAt      time.Time `json:"notified_at,omitzero"`
}

var semverPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

func IsReleaseVersion(v string) bool {
	return semverPattern.MatchString(v)
}

func ReleaseNotesURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", Repo, strings.TrimPrefix(version, "v"))
}

var noRedirectClient = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

func FetchLatest(ctx context.Context, timeout time.Duration) (*Release, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	base, err := url.Parse(downloadBaseURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadBaseURL+"/"+Repo+"/releases/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", projectName)

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
	default:
		return nil, fmt.Errorf("checking for updates: %s/releases/latest returned HTTP %d, not a redirect to the latest release", Repo, resp.StatusCode)
	}
	loc, err := resp.Location()
	if err != nil {
		return nil, fmt.Errorf("checking for updates: the latest-release redirect has no Location")
	}
	tagPrefix := "/" + Repo + "/releases/tag/"
	tag, ok := strings.CutPrefix(loc.Path, tagPrefix)
	if loc.Scheme != base.Scheme || loc.Host != base.Host || !ok {
		return nil, fmt.Errorf("checking for updates: unexpected latest-release redirect to %s", loc.Redacted())
	}
	if !strings.HasPrefix(tag, "v") || !IsReleaseVersion(tag) {
		return nil, fmt.Errorf("checking for updates: latest release has an unexpected tag %q", tag)
	}
	return &Release{Version: strings.TrimPrefix(tag, "v")}, nil
}

func CheckNow(ctx context.Context, current, configDir string) (*Info, error) {
	rel, err := FetchLatest(ctx, ForegroundTimeout)
	if err != nil {
		return nil, err
	}
	entry := loadCache(configDir)
	entry.LastChecked = now().UTC()
	entry.LatestVersion = rel.Version
	saveCache(configDir, entry)
	return newInfo(current, rel.Version), nil
}

func Cached(current, configDir string) (info *Info, ok bool) {
	entry := loadCache(configDir)
	if entry.LatestVersion == "" {
		return nil, false
	}
	return newInfo(current, entry.LatestVersion), true
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
