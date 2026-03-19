package node

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdNodeStats(f *cmdutil.Factory) *cobra.Command {
	var metric string

	cmd := &cobra.Command{
		Use:   "stats [node-id]",
		Short: "Show node statistics",
		Long: `Display statistics for one or all nodes.

Available metrics: indices, os, process, jvm, thread_pool, fs, transport, http, breaker.

Examples:
  # Show stats for all nodes
  es node stats

  # Show stats for a specific node
  es node stats node-1

  # Show only JVM stats
  es node stats --metric jvm

  # Output as JSON
  es node stats -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			nodeID := ""
			if len(args) > 0 {
				nodeID = args[0]
			}

			stats, err := c.GetNodeStats(context.Background(), nodeID, metric)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, stats, nil)
		},
	}

	cmd.Flags().StringVar(&metric, "metric", "", "Specific metric to retrieve (e.g., jvm, os, indices)")

	return cmd
}
