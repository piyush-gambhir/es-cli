package search

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdSearchFieldCaps(f *cmdutil.Factory) *cobra.Command {
	var fields string

	cmd := &cobra.Command{
		Use:   "field-caps <index>",
		Short: "Show field capabilities",
		Long: `Display field capabilities for the given index.

Use --fields to specify which fields to return capabilities for.

Examples:
  # Get capabilities for all fields
  es search field-caps my-index --fields "*"

  # Get capabilities for specific fields
  es search field-caps my-index --fields "timestamp,message,level"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fields == "" {
				return cmdutil.FlagErrorf("--fields is required")
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.FieldCaps(context.Background(), args[0], fields)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated list of fields")

	return cmd
}
