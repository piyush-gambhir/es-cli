package cluster

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdClusterSettings(f *cmdutil.Factory) *cobra.Command {
	var includeDefaults bool

	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Show cluster settings",
		Long: `Display the current cluster settings (persistent and transient).

By default, only explicitly configured settings are shown. Use --include-defaults
to also display the default values.

Examples:
  # Show cluster settings
  es cluster settings

  # Include default settings
  es cluster settings --include-defaults

  # Output as JSON
  es cluster settings -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			settings, err := c.GetClusterSettings(context.Background(), includeDefaults)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, settings, nil)
		},
	}

	cmd.Flags().BoolVar(&includeDefaults, "include-defaults", false, "Include default settings in the output")

	return cmd
}
