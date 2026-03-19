package index

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdIndexOpen(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "open <index>",
		Short: "Open a closed index",
		Long: `Open a previously closed Elasticsearch index, making it available for read and write operations.

Examples:
  # Open an index
  es index open my-index`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			name := args[0]

			if err := c.OpenIndex(context.Background(), name); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Index %s opened.\n", name)
			}
			return nil
		},
	}
}
