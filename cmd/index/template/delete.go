package template

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdTemplateDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		confirm  bool
		ifExists bool
	)

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an index template",
		Long: `Delete an index template.

Requires confirmation unless --confirm is provided.

Examples:
  # Delete a template (interactive confirmation)
  es index template delete my-template

  # Delete without confirmation
  es index template delete my-template --confirm

  # Delete only if it exists
  es index template delete my-template --confirm --if-exists`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ok, err := cmdutil.ConfirmAction(f.IOStreams.In, f.IOStreams.Out,
				fmt.Sprintf("Delete index template %s?", name), confirm, f.NoInput)
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

			err = c.DeleteIndexTemplate(context.Background(), name)
			if err != nil {
				if ifExists && client.IsNotFound(err) {
					if !f.Quiet {
						fmt.Fprintf(f.IOStreams.Out, "Index template %s does not exist.\n", name)
					}
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Index template %s deleted.\n", name)
			}
			return nil
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm)
	cmdutil.AddIfExistsFlag(cmd, &ifExists)

	return cmd
}
