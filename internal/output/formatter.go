package output

import (
	"fmt"
	"io"
)

// Format represents an output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// TableDef defines how to render data in a table.
type TableDef struct {
	Headers []string
	RowFunc func(item interface{}) []string
}

// Formatter is the interface for output formatters.
type Formatter interface {
	Format(w io.Writer, data interface{}, tableDef *TableDef) error
}

// NewFormatter creates a formatter for the given format string.
func NewFormatter(format string) (Formatter, error) {
	switch Format(format) {
	case FormatTable:
		return &TableFormatter{}, nil
	case FormatJSON:
		return &JSONFormatter{}, nil
	case FormatYAML:
		return &YAMLFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s (use table, json, or yaml)", format)
	}
}

// Print is a convenience function that formats data and writes to w.
func Print(w io.Writer, format string, data interface{}, tableDef *TableDef) error {
	f, err := NewFormatter(format)
	if err != nil {
		return err
	}
	return f.Format(w, data, tableDef)
}
