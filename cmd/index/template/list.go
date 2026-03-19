package template

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdTemplateList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List index templates",
		Long: `List all index templates in the cluster.

Shows template name, index patterns, order, and version.

Examples:
  # List all templates
  es index template list

  # Output as JSON
  es index template list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			templates, err := c.ListIndexTemplates(context.Background())
			if err != nil {
				return err
			}

			if len(templates) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No templates found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, templates, &output.TableDef{
				Headers: []string{"Name", "Index Patterns", "Order", "Version"},
				RowFunc: func(item interface{}) []string {
					t := item.(client.CatTemplate)
					return []string{
						t.Name,
						t.IndexPatterns,
						t.Order,
						t.Version,
					}
				},
			})
		},
	}
}
