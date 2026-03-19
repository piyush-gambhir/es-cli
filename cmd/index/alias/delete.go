package alias

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdAliasDelete(f *cmdutil.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <index> <alias>",
		Short: "Delete an alias",
		Long: `Remove an alias from an Elasticsearch index.

Requires confirmation unless --confirm is provided.

Examples:
  # Delete an alias (interactive confirmation)
  es index alias delete my-index my-alias

  # Delete without confirmation
  es index alias delete my-index my-alias --confirm`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			index := args[0]
			alias := args[1]

			ok, err := cmdutil.ConfirmAction(f.IOStreams.In, f.IOStreams.Out,
				fmt.Sprintf("Delete alias %s from index %s?", alias, index), confirm, f.NoInput)
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

			if err := c.DeleteAlias(context.Background(), index, alias); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Alias %s deleted from index %s.\n", alias, index)
			}
			return nil
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm)

	return cmd
}
