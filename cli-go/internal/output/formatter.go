package output

import (
	"fmt"
	"io"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

type TableDef struct {
	Headers []string
	RowFunc func(item interface{}) []string
}

type Formatter interface {
	Format(w io.Writer, data interface{}, tableDef *TableDef) error
}

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

func Print(w io.Writer, format string, data interface{}, tableDef *TableDef) error {
	f, err := NewFormatter(format)
	if err != nil {
		return err
	}
	return f.Format(w, data, tableDef)
}
