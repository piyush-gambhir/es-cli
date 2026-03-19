package index

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdIndexDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		confirm  bool
		ifExists bool
	)

	cmd := &cobra.Command{
		Use:   "delete <index>",
		Short: "Delete an index",
		Long: `Delete an Elasticsearch index.

This is a destructive operation that permanently removes the index and all its data.
Requires confirmation unless --confirm is provided.

Examples:
  # Delete an index (interactive confirmation)
  es index delete my-index

  # Delete without confirmation
  es index delete my-index --confirm

  # Delete only if it exists
  es index delete my-index --confirm --if-exists`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ok, err := cmdutil.ConfirmAction(f.IOStreams.In, f.IOStreams.Out,
				fmt.Sprintf("Delete index %s?", name), confirm, f.NoInput)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Fprintln(f.IOStreams.Out, "Aborted.")
				return nil
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			err = c.DeleteIndex(context.Background(), name)
			if err != nil {
				if ifExists && client.IsNotFound(err) {
					if !f.Quiet {
						fmt.Fprintf(f.IOStreams.Out, "Index %s does not exist.\n", name)
					}
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Index %s deleted.\n", name)
			}
			return nil
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm)
	cmdutil.AddIfExistsFlag(cmd, &ifExists)

	return cmd
}
