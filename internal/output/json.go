package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONFormatter outputs data as pretty-printed JSON.
type JSONFormatter struct{}

// Format writes data as indented JSON.
func (f *JSONFormatter) Format(w io.Writer, data interface{}, _ *TableDef) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}
