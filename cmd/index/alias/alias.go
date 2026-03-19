package alias

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdAlias returns the alias parent command.
func NewCmdAlias(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alias",
		Short:   "Manage index aliases",
		Long:    "Create, delete, and list index aliases.",
		Aliases: []string{"aliases"},
	}

	cmd.AddCommand(newCmdAliasList(f))
	cmd.AddCommand(newCmdAliasCreate(f))
	cmd.AddCommand(newCmdAliasDelete(f))

	return cmd
}
