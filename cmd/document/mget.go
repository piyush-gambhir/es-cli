package document

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdDocumentMGet(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "mget <index>",
		Short: "Multi-get documents",
		Long: `Retrieve multiple documents in a single request.

Provide the document IDs body via a JSON/YAML file with -f.

Examples:
  # Multi-get from file
  es document mget my-index -f ids.json

  # Multi-get from stdin
  echo '{"ids":["1","2","3"]}' | es doc mget my-index -f -`,
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

			result, err := c.MultiGet(context.Background(), args[0], body)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
