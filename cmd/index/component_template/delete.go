package componenttemplate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdComponentTemplateDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		confirm  bool
		ifExists bool
	)

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a component template",
		Long: `Delete a component template.

Requires confirmation unless --confirm is provided.

Examples:
  # Delete a component template (interactive confirmation)
  es index component-template delete my-component

  # Delete without confirmation
  es index component-template delete my-component --confirm

  # Delete only if it exists
  es index component-template delete my-component --confirm --if-exists`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ok, err := cmdutil.ConfirmAction(f.IOStreams.In, f.IOStreams.Out,
				fmt.Sprintf("Delete component template %s?", name), confirm, f.NoInput)
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

			err = c.DeleteComponentTemplate(context.Background(), name)
			if err != nil {
				if ifExists && client.IsNotFound(err) {
					if !f.Quiet {
						fmt.Fprintf(f.IOStreams.Out, "Component template %s does not exist.\n", name)
					}
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Component template %s deleted.\n", name)
			}
			return nil
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm)
	cmdutil.AddIfExistsFlag(cmd, &ifExists)

	return cmd
}
