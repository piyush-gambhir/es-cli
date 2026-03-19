package index

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIndexRollover(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "rollover <alias>",
		Short: "Rollover an index alias",
		Long: `Rollover an index alias to a new index based on conditions.

Optionally provide rollover conditions via a JSON/YAML file with -f.

Examples:
  # Rollover unconditionally
  es index rollover my-alias

  # Rollover with conditions from a file
  es index rollover my-alias -f conditions.json`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			var body interface{}
			if file != "" {
				if err := cmdutil.UnmarshalInput(file, &body); err != nil {
					return err
				}
			}

			result, err := c.Rollover(context.Background(), args[0], body)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
