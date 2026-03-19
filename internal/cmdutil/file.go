package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReadInput reads data from a file path or stdin (if path is "-").
func ReadInput(path string) ([]byte, error) {
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		return data, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}
	return data, nil
}

// UnmarshalInput reads from the given path and unmarshals into v.
// Automatically detects JSON vs YAML based on content.
func UnmarshalInput(path string, v interface{}) error {
	data, err := ReadInput(path)
	if err != nil {
		return err
	}

	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("parsing JSON: %w", err)
		}
		return nil
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}
	return nil
}
