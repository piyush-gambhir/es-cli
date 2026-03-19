package ingest

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIngestList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ingest pipelines",
		Long: `List all ingest pipelines in the cluster.

Examples:
  # List all pipelines
  es ingest list

  # Output as JSON
  es ingest list -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.ListPipelines(context.Background())
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	return cmd
}
