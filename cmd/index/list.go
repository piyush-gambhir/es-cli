package index

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIndexList(f *cmdutil.Factory) *cobra.Command {
	var (
		pattern string
		health  string
		status  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List indices",
		Long: `List all indices in the cluster.

Shows index name, health, status, UUID, primary/replica counts, document count,
and store size. Use flags to filter by pattern, health, or status.

Examples:
  # List all indices
  es index list

  # Filter by pattern
  es index list --pattern "my-index-*"

  # Filter by health
  es index list --health yellow

  # Filter by status
  es index list --status open

  # Output as JSON
  es index list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			indices, err := c.ListIndices(context.Background(), pattern, health, status)
			if err != nil {
				return err
			}

			if len(indices) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No indices found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, indices, &output.TableDef{
				Headers: []string{"Index", "Health", "Status", "UUID", "Pri", "Rep", "Docs Count", "Store Size"},
				RowFunc: func(item interface{}) []string {
					i := item.(client.CatIndex)
					return []string{
						i.Index,
						i.Health,
						i.Status,
						i.UUID,
						i.Pri,
						i.Rep,
						i.DocsCount,
						i.StoreSize,
					}
				},
			})
		},
	}

	cmd.Flags().StringVar(&pattern, "pattern", "", "Index name pattern (supports wildcards)")
	cmd.Flags().StringVar(&health, "health", "", "Filter by health: green, yellow, red")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: open, close")

	return cmd
}
