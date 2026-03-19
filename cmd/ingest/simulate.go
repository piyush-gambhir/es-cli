package ingest

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIngestSimulate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "simulate <name>",
		Short: "Simulate an ingest pipeline",
		Long: `Simulate an ingest pipeline with sample documents.

Provide the documents body via a JSON/YAML file with -f.

Examples:
  # Simulate a pipeline
  es ingest simulate my-pipeline -f docs.json

  # Simulate from stdin
  echo '{"docs":[{"_source":{"message":"test"}}]}' | es ingest simulate my-pipeline -f -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return cmdutil.FlagErrorf("file is required (-f)")
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			var body interface{}
			if err := cmdutil.UnmarshalInput(file, &body); err != nil {
				return err
			}

			result, err := c.SimulatePipeline(context.Background(), args[0], body)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
