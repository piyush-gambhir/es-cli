package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter outputs data as YAML.
type YAMLFormatter struct{}

// Format writes data as YAML.
//
// Data is normalized through JSON first so that json.RawMessage payloads
// (returned by most API calls) render as structured YAML instead of raw byte
// sequences, and struct fields use their JSON tag names, consistent with the
// json output format.
func (f *YAMLFormatter) Format(w io.Writer, data interface{}, _ *TableDef) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}

	// JSON is valid YAML; unmarshaling with yaml.v3 (rather than encoding/json)
	// preserves large integers instead of converting them to float64.
	var normalized interface{}
	if err := yaml.Unmarshal(jsonBytes, &normalized); err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}

	out, err := yaml.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	_, err = fmt.Fprint(w, string(out))
	return err
}
