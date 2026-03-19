package node

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdNode returns the node parent command.
func NewCmdNode(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Short:   "Manage cluster nodes",
		Long:    "View node information, statistics, and hot threads.",
		Aliases: []string{"nodes"},
	}

	cmd.AddCommand(newCmdNodeList(f))
	cmd.AddCommand(newCmdNodeInfo(f))
	cmd.AddCommand(newCmdNodeStats(f))
	cmd.AddCommand(newCmdNodeHotThreads(f))

	return cmd
}
