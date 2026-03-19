package cluster

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdClusterHealth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Show cluster health status",
		Long: `Display the current health status of the Elasticsearch cluster.

Shows the cluster name, status (green/yellow/red), number of nodes,
active shards, and other health indicators.

Examples:
  # Show cluster health
  es cluster health

  # Output as JSON
  es cluster health -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			health, err := c.GetClusterHealth(context.Background())
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, health, &output.TableDef{
				Headers: []string{"Cluster", "Status", "Nodes", "Data Nodes", "Active Shards", "Unassigned", "Relocating"},
				RowFunc: func(item interface{}) []string {
					h := item.(*client.ClusterHealth)
					return []string{
						h.ClusterName,
						h.Status,
						fmt.Sprintf("%d", h.NumberOfNodes),
						fmt.Sprintf("%d", h.NumberOfDataNodes),
						fmt.Sprintf("%d", h.ActiveShards),
						fmt.Sprintf("%d", h.UnassignedShards),
						fmt.Sprintf("%d", h.RelocatingShards),
					}
				},
			})
		},
	}

	return cmd
}
