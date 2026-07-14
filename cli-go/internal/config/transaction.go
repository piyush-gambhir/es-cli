package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/flock"
)

var updateMu sync.Mutex

// Update performs a cross-process locked load-modify-save transaction.
func Update(mutator func(*Config) error) error {
	updateMu.Lock()
	defer updateMu.Unlock()
	path := ConfigFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	fileLock := flock.New(path + ".lock")
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("locking config: %w", err)
	}
	defer fileLock.Unlock()
	if err := os.Chmod(path+".lock", 0o600); err != nil {
		return fmt.Errorf("securing config lock: %w", err)
	}
	cfg, err := LoadFrom(path)
	if err != nil {
		return err
	}
	if err := mutator(cfg); err != nil {
		return err
	}
	return cfg.SaveTo(path)
}
