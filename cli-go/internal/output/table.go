package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

// TableFormatter outputs data as a formatted ASCII table.
type TableFormatter struct{}

// Format writes data as a table. If tableDef is nil, falls back to JSON.
func (f *TableFormatter) Format(w io.Writer, data interface{}, tableDef *TableDef) error {
	if tableDef == nil {
		// Fall back to JSON if no table definition.
		return (&JSONFormatter{}).Format(w, data, nil)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Print headers in uppercase.
	headers := make([]string, len(tableDef.Headers))
	for i, h := range tableDef.Headers {
		headers[i] = strings.ToUpper(h)
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	// Handle both slices and single items.
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			row := tableDef.RowFunc(item)
			fmt.Fprintln(tw, strings.Join(row, "\t"))
		}
	default:
		row := tableDef.RowFunc(data)
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	return tw.Flush()
}

// PrintMessage writes a simple message to the writer.
func PrintMessage(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+"\n", args...)
}
