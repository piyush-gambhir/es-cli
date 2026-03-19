package node

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdNodeList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List cluster nodes",
		Long: `List all nodes in the Elasticsearch cluster.

Shows IP address, heap/RAM usage, CPU, load averages, roles, and master status.

Examples:
  # List all nodes
  es node list

  # Output as JSON
  es node list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			nodes, err := c.ListNodes(context.Background())
			if err != nil {
				return err
			}

			if len(nodes) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No nodes found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, nodes, &output.TableDef{
				Headers: []string{"Name", "IP", "Heap%", "RAM%", "CPU", "Load 1m", "Role", "Master"},
				RowFunc: func(item interface{}) []string {
					n := item.(client.CatNode)
					return []string{
						n.Name,
						n.IP,
						n.HeapPercent,
						n.RAMPercent,
						n.CPU,
						n.Load1m,
						n.NodeRole,
						n.Master,
					}
				},
			})
		},
	}
}
