package cmdutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

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

func UnmarshalInput(path string, v interface{}) error {
	data, err := ReadInput(path)
	if err != nil {
		return err
	}

	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.UseNumber()
		if err := dec.Decode(v); err != nil {
			return fmt.Errorf("parsing JSON: %w", err)
		}
		if _, err := dec.Token(); err != io.EOF {
			return fmt.Errorf("parsing JSON: unexpected data after top-level value")
		}
		return nil
	}

	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}
	return nil
}
