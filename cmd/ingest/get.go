package ingest

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIngestGet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get an ingest pipeline",
		Long: `Get the definition of an ingest pipeline by name.

Examples:
  # Get a pipeline
  es ingest get my-pipeline

  # Output as JSON
  es ingest get my-pipeline -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.GetPipeline(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	return cmd
}
