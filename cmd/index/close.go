package index

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdIndexClose(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "close <index>",
		Short: "Close an index",
		Long: `Close an Elasticsearch index, blocking read and write operations.

A closed index consumes minimal cluster resources. Use "es index open" to reopen it.

Examples:
  # Close an index
  es index close my-index`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			name := args[0]

			if err := c.CloseIndex(context.Background(), name); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Index %s closed.\n", name)
			}
			return nil
		},
	}
}
