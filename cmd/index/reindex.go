package index

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdReindex(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Reindex documents from one index to another",
		Long: `Copy documents from a source index to a destination index.

Requires a reindex body provided via -f with source and dest configuration.

Examples:
  # Reindex from a file
  es index reindex -f reindex.json

  # Reindex from stdin
  cat reindex.json | es index reindex -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return cmdutil.FlagErrorf("required flag -f not set")
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			var body interface{}
			if err := cmdutil.UnmarshalInput(file, &body); err != nil {
				return err
			}

			result, err := c.Reindex(context.Background(), body)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
