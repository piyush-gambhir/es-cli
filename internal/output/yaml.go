package output

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter outputs data as YAML.
type YAMLFormatter struct{}

// Format writes data as YAML.
func (f *YAMLFormatter) Format(w io.Writer, data interface{}, _ *TableDef) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	_, err = fmt.Fprint(w, string(out))
	return err
}
