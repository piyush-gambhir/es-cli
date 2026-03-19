package template

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdTemplate returns the template parent command.
func NewCmdTemplate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "Manage index templates",
		Long:    "Create, delete, and list index templates.",
		Aliases: []string{"templates", "tmpl"},
	}

	cmd.AddCommand(newCmdTemplateList(f))
	cmd.AddCommand(newCmdTemplateGet(f))
	cmd.AddCommand(newCmdTemplateCreate(f))
	cmd.AddCommand(newCmdTemplateDelete(f))

	return cmd
}
