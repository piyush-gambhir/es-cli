package search

import (
	"bytes"
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdSearchMSearch(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "msearch",
		Short: "Execute a multi-search request",
		Long: `Execute multiple search requests in a single API call using NDJSON format.

The file must contain alternating header and body lines in NDJSON format.

Examples:
  # Multi-search from file
  es search msearch -f requests.ndjson

  # Multi-search from stdin
  cat requests.ndjson | es search msearch -f -`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return cmdutil.FlagErrorf("file is required (-f)")
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			data, err := cmdutil.ReadInput(file)
			if err != nil {
				return err
			}

			result, err := c.MultiSearch(context.Background(), bytes.NewReader(data))
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
