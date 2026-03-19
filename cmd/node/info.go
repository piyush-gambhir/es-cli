package node

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdNodeInfo(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "info <node-id>",
		Short: "Show detailed node information",
		Long: `Display detailed information about a specific node.

Examples:
  # Show node info by node ID or name
  es node info node-1

  # Output as JSON
  es node info node-1 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			info, err := c.GetNodeInfo(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, info, nil)
		},
	}
}
