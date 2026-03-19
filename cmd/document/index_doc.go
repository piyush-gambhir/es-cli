package document

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdDocumentIndex(f *cmdutil.Factory) *cobra.Command {
	var (
		file string
		id   string
	)

	cmd := &cobra.Command{
		Use:   "index <index>",
		Short: "Index (create or update) a document",
		Long: `Index a document into an Elasticsearch index.

If --id is provided, the document is indexed with that ID (PUT).
Otherwise, an ID is auto-generated (POST).

Examples:
  # Index a document with auto-generated ID
  es document index my-index -f doc.json

  # Index a document with a specific ID
  es document index my-index -f doc.json --id abc123

  # Index from stdin
  echo '{"name":"test"}' | es doc index my-index -f -`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"mutates": "true"},
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

			result, err := c.IndexDocument(context.Background(), args[0], id, body)
			if err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.ErrOut, "Document indexed in %s\n", args[0])
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)
	cmd.Flags().StringVar(&id, "id", "", "Document ID (auto-generated if not set)")

	return cmd
}
