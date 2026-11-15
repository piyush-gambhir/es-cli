package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

type Profile struct {
	URL        string `yaml:"url,omitempty"`
	AuthMethod string `yaml:"auth_method,omitempty"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
	APIKeyID   string `yaml:"api_key_id,omitempty"`
	APIKey     string `yaml:"api_key,omitempty"`
	Token      string `yaml:"token,omitempty"`
	CACert     string `yaml:"ca_cert,omitempty"`
	Insecure   bool   `yaml:"insecure,omitempty"`
	ReadOnly   bool   `yaml:"read_only,omitempty"`
}

type Defaults struct {
	Output string `yaml:"output,omitempty"`
}

type Config struct {
	CurrentProfile string             `yaml:"current_profile,omitempty"`
	Profiles       map[string]Profile `yaml:"profiles,omitempty"`
	Defaults       Defaults           `yaml:"defaults,omitempty"`
}

func Load() (*Config, error) {
	return LoadFrom(ConfigFilePath())
}

func LoadFrom(path string) (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
		Defaults: Defaults{Output: "table"},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	if cfg.Defaults.Output == "" {
		cfg.Defaults.Output = "table"
	}

	return cfg, nil
}

func (c *Config) Save() error {
	return c.SaveTo(ConfigFilePath())
}

func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("securing config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := atomicWriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("securing config: %w", err)
	}

	return nil
}

func (c *Config) CurrentProfileConfig() *Profile {
	if c.CurrentProfile == "" {
		return nil
	}
	p, ok := c.Profiles[c.CurrentProfile]
	if !ok {
		return nil
	}
	return &p
}

func (c *Config) CreateProfile(name string, profile Profile) error {
	if _, exists := c.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	c.Profiles[name] = profile
	return nil
}

func (c *Config) DeleteProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(c.Profiles, name)
	if c.CurrentProfile == name {
		c.CurrentProfile = ""
	}
	return nil
}

func (c *Config) SetCurrentProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}
	c.CurrentProfile = name
	return nil
}
