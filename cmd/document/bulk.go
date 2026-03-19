package document

import (
	"bytes"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdDocumentBulk(f *cmdutil.Factory) *cobra.Command {
	var (
		file  string
		index string
	)

	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk index documents",
		Long: `Bulk index documents using NDJSON format.

The file must be in NDJSON (newline-delimited JSON) format with alternating
action and source lines.

Examples:
  # Bulk index from file
  es document bulk -f bulk.ndjson

  # Bulk index into a specific index
  es document bulk -f bulk.ndjson --index my-index

  # Bulk index from stdin
  cat bulk.ndjson | es doc bulk -f -`,
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"mutates": "true"},
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

			result, err := c.BulkIndex(context.Background(), index, bytes.NewReader(data))
			if err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.ErrOut, "Bulk operation completed\n")
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)
	cmd.Flags().StringVar(&index, "index", "", "Default index for bulk operations")

	return cmd
}
