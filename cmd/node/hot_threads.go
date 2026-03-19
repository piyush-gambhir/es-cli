package node

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdNodeHotThreads(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "hot-threads [node-id]",
		Short: "Show hot threads on nodes",
		Long: `Display the current hot threads on one or all nodes.

Hot threads are threads that are consuming significant CPU time. This output
is always plain text, regardless of the --output flag.

Examples:
  # Show hot threads for all nodes
  es node hot-threads

  # Show hot threads for a specific node
  es node hot-threads node-1`,
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

			result, err := c.GetHotThreads(context.Background(), nodeID)
			if err != nil {
				return err
			}

			fmt.Fprint(f.IOStreams.Out, result)
			return nil
		},
	}
}
