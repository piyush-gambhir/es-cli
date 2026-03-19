package index

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIndexStats(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "stats <index>",
		Short: "Get index statistics",
		Long: `Display statistics for an Elasticsearch index.

Includes document counts, store size, indexing rates, search rates, and more.

Examples:
  # Get index stats
  es index stats my-index

  # Output as JSON
  es index stats my-index -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.GetIndexStats(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}
}
