package cluster

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdCluster returns the cluster parent command.
func NewCmdCluster(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage the Elasticsearch cluster",
		Long:  "View cluster health, stats, settings, pending tasks, and allocation information.",
	}

	cmd.AddCommand(newCmdClusterHealth(f))
	cmd.AddCommand(newCmdClusterStats(f))
	cmd.AddCommand(newCmdClusterSettings(f))
	cmd.AddCommand(newCmdClusterPendingTasks(f))
	cmd.AddCommand(newCmdClusterAllocationExplain(f))

	return cmd
}
