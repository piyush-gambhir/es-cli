package cluster

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdClusterStats(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show cluster statistics",
		Long: `Display detailed statistics about the Elasticsearch cluster.

Includes information about nodes, indices, shards, storage, and more.

Examples:
  # Show cluster stats
  es cluster stats

  # Output as JSON for parsing
  es cluster stats -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			stats, err := c.GetClusterStats(context.Background())
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, stats, nil)
		},
	}
}
