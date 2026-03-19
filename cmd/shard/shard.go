package shard

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

// NewCmdShard returns the shard parent command.
func NewCmdShard(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shard",
		Short:   "View shard information",
		Long:    "View shard allocation and status information.",
		Aliases: []string{"shards"},
	}

	cmd.AddCommand(newCmdShardList(f))

	return cmd
}

func newCmdShardList(f *cmdutil.Factory) *cobra.Command {
	var index string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shards",
		Long: `List all shards in the cluster, optionally filtered by index.

Shows index, shard number, primary/replica status, state, document count,
store size, IP, and node name.

Examples:
  # List all shards
  es shard list

  # Filter by index
  es shard list --index my-index

  # Output as JSON
  es shard list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			shards, err := c.ListShards(context.Background(), index)
			if err != nil {
				return err
			}

			if len(shards) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No shards found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, shards, &output.TableDef{
				Headers: []string{"Index", "Shard", "Pri/Rep", "State", "Docs", "Store", "IP", "Node"},
				RowFunc: func(item interface{}) []string {
					s := item.(client.CatShard)
					return []string{
						s.Index,
						s.Shard,
						s.PriRep,
						s.State,
						s.Docs,
						s.Store,
						s.IP,
						s.Node,
					}
				},
			})
		},
	}

	cmd.Flags().StringVar(&index, "index", "", "Filter shards by index name")

	return cmd
}
