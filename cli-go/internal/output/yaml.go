package output

import (
	"encoding/json"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(w io.Writer, data interface{}, _ *TableDef) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}

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
