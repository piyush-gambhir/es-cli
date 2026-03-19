package ilm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdILMDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		confirm  bool
		ifExists bool
	)

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an ILM policy",
		Long: `Delete an ILM policy by name.

Requires confirmation unless --confirm is provided.

Examples:
  # Delete a policy (with confirmation prompt)
  es ilm delete my-policy

  # Delete without confirmation
  es ilm delete my-policy --confirm

  # Ignore if policy does not exist
  es ilm delete my-policy --confirm --if-exists`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"mutates": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := cmdutil.ConfirmAction(f.IOStreams.In, f.IOStreams.ErrOut,
				fmt.Sprintf("Delete ILM policy %s?", args[0]),
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

			err = c.DeleteILMPolicy(context.Background(), args[0])
			if err != nil {
				if ifExists && client.IsNotFound(err) {
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.ErrOut, "ILM policy %s deleted\n", args[0])
			}

			return nil
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm)
	cmdutil.AddIfExistsFlag(cmd, &ifExists)

	return cmd
}
