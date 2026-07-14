package template

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/output"
)

func newCmdTemplateGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get an index template",
		Long: `Display the full definition of an index template.

Examples:
  # Get a template
  es index template get my-template

  # Output as JSON
  es index template get my-template -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.GetIndexTemplate(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}
}
