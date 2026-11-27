package config

import (
	"os"
	"strings"
)

type ResolvedConfig struct {
	URL        string
	AuthMethod string
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

func Resolve(flagURL, flagUsername, flagPassword, flagAPIKeyID, flagAPIKey, flagToken, flagCACert string, flagInsecure bool, profile *Profile, defaults Defaults) *ResolvedConfig {
	rc := &ResolvedConfig{}

	rc.URL = firstNonEmpty(flagURL, os.Getenv("ES_URL"))
	if rc.URL == "" && profile != nil {
		rc.URL = profile.URL
	}

	rc.Username = firstNonEmpty(flagUsername, os.Getenv("ES_USERNAME"))
	if rc.Username == "" && profile != nil {
		rc.Username = profile.Username
	}

	rc.Password = firstNonEmpty(flagPassword, os.Getenv("ES_PASSWORD"))
	if rc.Password == "" && profile != nil {
		rc.Password = profile.Password
	}

	rc.APIKeyID = firstNonEmpty(flagAPIKeyID, os.Getenv("ES_API_KEY_ID"))
	if rc.APIKeyID == "" && profile != nil {
		rc.APIKeyID = profile.APIKeyID
	}

	rc.APIKey = firstNonEmpty(flagAPIKey, os.Getenv("ES_API_KEY"))
	if rc.APIKey == "" && profile != nil {
		rc.APIKey = profile.APIKey
	}

	rc.Token = firstNonEmpty(flagToken, os.Getenv("ES_TOKEN"))
	if rc.Token == "" && profile != nil {
		rc.Token = profile.Token
	}

	rc.CACert = firstNonEmpty(flagCACert, os.Getenv("ES_CA_CERT"))
	if rc.CACert == "" && profile != nil {
		rc.CACert = profile.CACert
	}

	if flagInsecure {
		rc.Insecure = true
	} else if envInsecure := os.Getenv("ES_INSECURE"); envInsecure != "" {
		rc.Insecure = strings.EqualFold(envInsecure, "true") || envInsecure == "1"
	} else if profile != nil {
		rc.Insecure = profile.Insecure
	}

	rc.AuthMethod = detectAuthMethod(&ResolvedConfig{
		Username: firstNonEmpty(flagUsername, os.Getenv("ES_USERNAME")),
		Password: firstNonEmpty(flagPassword, os.Getenv("ES_PASSWORD")),
		APIKeyID: firstNonEmpty(flagAPIKeyID, os.Getenv("ES_API_KEY_ID")),
		APIKey:   firstNonEmpty(flagAPIKey, os.Getenv("ES_API_KEY")),
		Token:    firstNonEmpty(flagToken, os.Getenv("ES_TOKEN")),
	})
	if rc.AuthMethod == "" && profile != nil {
		rc.AuthMethod = profile.AuthMethod
	}
	if rc.AuthMethod == "" {
		rc.AuthMethod = detectAuthMethod(rc)
	}

	rc.Output = defaults.Output
	if rc.Output == "" {
		rc.Output = "table"
	}

	if profile != nil {
		rc.ReadOnly = profile.ReadOnly
	}
	if envRO := os.Getenv("ES_READ_ONLY"); envRO != "" {
		rc.ReadOnly = strings.EqualFold(envRO, "true") || envRO == "1"
	}

	return rc
}

func detectAuthMethod(rc *ResolvedConfig) string {
	if rc.Token != "" {
		return "bearer"
	}
	if rc.APIKeyID != "" && rc.APIKey != "" {
		return "api_key"
	}
	if rc.APIKey != "" {
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
