package config

import (
	"os"
	"strings"
)

// ResolvedConfig holds the final resolved configuration after layering
// flags > env vars > profile values.
type ResolvedConfig struct {
	URL        string
	AuthMethod string // basic, api_key, bearer, or "" (auto-detect)
	Username   string
	Password   string
	APIKeyID   string
	APIKey     string
	Token      string
	CACert     string
	Insecure   bool
	Output     string
	ReadOnly   bool
}

// Resolve merges flag values, environment variables, and profile values.
// Priority: flags > env vars > profile values.
func Resolve(flagURL, flagUsername, flagPassword, flagAPIKeyID, flagAPIKey, flagToken, flagCACert string, flagInsecure bool, profile *Profile, defaults Defaults) *ResolvedConfig {
	rc := &ResolvedConfig{}

	// URL
	rc.URL = firstNonEmpty(flagURL, os.Getenv("ES_URL"))
	if rc.URL == "" && profile != nil {
		rc.URL = profile.URL
	}

	// Username
	rc.Username = firstNonEmpty(flagUsername, os.Getenv("ES_USERNAME"))
	if rc.Username == "" && profile != nil {
		rc.Username = profile.Username
	}

	// Password
	rc.Password = firstNonEmpty(flagPassword, os.Getenv("ES_PASSWORD"))
	if rc.Password == "" && profile != nil {
		rc.Password = profile.Password
	}

	// API Key ID
	rc.APIKeyID = firstNonEmpty(flagAPIKeyID, os.Getenv("ES_API_KEY_ID"))
	if rc.APIKeyID == "" && profile != nil {
		rc.APIKeyID = profile.APIKeyID
	}

	// API Key
	rc.APIKey = firstNonEmpty(flagAPIKey, os.Getenv("ES_API_KEY"))
	if rc.APIKey == "" && profile != nil {
		rc.APIKey = profile.APIKey
	}

	// Token
	rc.Token = firstNonEmpty(flagToken, os.Getenv("ES_TOKEN"))
	if rc.Token == "" && profile != nil {
		rc.Token = profile.Token
	}

	// CA Cert
	rc.CACert = firstNonEmpty(flagCACert, os.Getenv("ES_CA_CERT"))
	if rc.CACert == "" && profile != nil {
		rc.CACert = profile.CACert
	}

	// Insecure
	if flagInsecure {
		rc.Insecure = true
	} else if envInsecure := os.Getenv("ES_INSECURE"); envInsecure != "" {
		rc.Insecure = strings.EqualFold(envInsecure, "true") || envInsecure == "1"
	} else if profile != nil {
		rc.Insecure = profile.Insecure
	}

	// Auth method: profile value, then auto-detect from credentials.
	if profile != nil {
		rc.AuthMethod = profile.AuthMethod
	}
	if rc.AuthMethod == "" {
		rc.AuthMethod = detectAuthMethod(rc)
	}

	// Output
	rc.Output = defaults.Output
	if rc.Output == "" {
		rc.Output = "table"
	}

	// ReadOnly: profile value first, then env var overrides.
	if profile != nil {
		rc.ReadOnly = profile.ReadOnly
	}
	if envRO := os.Getenv("ES_READ_ONLY"); envRO != "" {
		rc.ReadOnly = strings.EqualFold(envRO, "true") || envRO == "1"
	}

	return rc
}

// detectAuthMethod auto-detects the auth method from available credentials.
func detectAuthMethod(rc *ResolvedConfig) string {
	if rc.Token != "" {
		return "bearer"
	}
	if rc.APIKeyID != "" && rc.APIKey != "" {
		return "api_key"
	}
	if rc.Username != "" && rc.Password != "" {
		return "basic"
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
