package componenttemplate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdComponentTemplateGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get a component template",
		Long: `Display the full definition of a component template.

Examples:
  # Get a component template
  es index component-template get my-component

  # Output as JSON
  es index component-template get my-component -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.GetComponentTemplate(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}
}
