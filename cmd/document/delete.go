package document

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdDocumentDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		confirm  bool
		ifExists bool
	)

	cmd := &cobra.Command{
		Use:   "delete <index> <id>",
		Short: "Delete a document by ID",
		Long: `Delete a single document from an index by its ID.

Requires confirmation unless --confirm is provided.

Examples:
  # Delete a document (with confirmation prompt)
  es document delete my-index abc123

  # Delete without confirmation
  es document delete my-index abc123 --confirm

  # Ignore if document does not exist
  es document delete my-index abc123 --confirm --if-exists`,
		Args:        cobra.ExactArgs(2),
		Annotations: map[string]string{"mutates": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := cmdutil.ConfirmAction(f.IOStreams.In, f.IOStreams.ErrOut,
				fmt.Sprintf("Delete document %s from index %s?", args[1], args[0]),
				confirm, f.NoInput)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			err = c.DeleteDocument(context.Background(), args[0], args[1])
			if err != nil {
				if ifExists && client.IsNotFound(err) {
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.ErrOut, "Document %s deleted from %s\n", args[1], args[0])
			}

			return nil
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm)
	cmdutil.AddIfExistsFlag(cmd, &ifExists)

	return cmd
}
