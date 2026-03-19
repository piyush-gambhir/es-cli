package index

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIndexGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <index>",
		Short: "Get index details",
		Long: `Display the full definition of an index including settings, mappings, and aliases.

Examples:
  # Get index details
  es index get my-index

  # Output as JSON
  es index get my-index -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.GetIndex(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}
}
