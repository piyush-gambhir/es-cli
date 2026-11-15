package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

type TableFormatter struct{}

func (f *TableFormatter) Format(w io.Writer, data interface{}, tableDef *TableDef) error {
	if tableDef == nil {
		return (&JSONFormatter{}).Format(w, data, nil)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	headers := make([]string, len(tableDef.Headers))
	for i, h := range tableDef.Headers {
		headers[i] = strings.ToUpper(h)
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

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

func PrintMessage(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+"\n", args...)
}
