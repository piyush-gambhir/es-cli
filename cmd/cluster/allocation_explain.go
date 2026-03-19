package cluster

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdClusterAllocationExplain(f *cmdutil.Factory) *cobra.Command {
	var (
		index   string
		shard   int
		primary bool
	)

	cmd := &cobra.Command{
		Use:   "allocation-explain",
		Short: "Explain shard allocation decisions",
		Long: `Provide an explanation for why a shard is unassigned or cannot be allocated.

Without arguments, explains the first unassigned shard found.
Use --index, --shard, and --primary to explain a specific shard.

Examples:
  # Explain first unassigned shard
  es cluster allocation-explain

  # Explain a specific shard
  es cluster allocation-explain --index my-index --shard 0 --primary

  # Output as JSON
  es cluster allocation-explain -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			var body interface{}
			if index != "" {
				body = map[string]interface{}{
					"index":   index,
					"shard":   shard,
					"primary": primary,
				}
			}

			result, err := c.AllocationExplain(context.Background(), body)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmd.Flags().StringVar(&index, "index", "", "Index name")
	cmd.Flags().IntVar(&shard, "shard", 0, "Shard number")
	cmd.Flags().BoolVar(&primary, "primary", false, "Explain primary shard (default: replica)")

	return cmd
}
