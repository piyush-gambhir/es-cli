package alias

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdAliasList(f *cmdutil.Factory) *cobra.Command {
	var index string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List aliases",
		Long: `List all index aliases in the cluster.

Optionally filter by index name.

Examples:
  # List all aliases
  es index alias list

  # Filter by index
  es index alias list --index my-index

  # Output as JSON
  es index alias list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			aliases, err := c.ListAliases(context.Background(), index)
			if err != nil {
				return err
			}

			if len(aliases) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No aliases found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, aliases, &output.TableDef{
				Headers: []string{"Alias", "Index", "Filter", "Is Write Index"},
				RowFunc: func(item interface{}) []string {
					a := item.(client.CatAlias)
					return []string{
						a.Alias,
						a.Index,
						a.Filter,
						a.IsWriteIndex,
					}
				},
			})
		},
	}

	cmd.Flags().StringVar(&index, "index", "", "Filter aliases by index name")

	return cmd
}
