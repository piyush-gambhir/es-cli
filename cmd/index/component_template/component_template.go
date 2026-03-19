package componenttemplate

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdComponentTemplate returns the component_template parent command.
func NewCmdComponentTemplate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "component-template",
		Short:   "Manage component templates",
		Long:    "Create, delete, and list component templates.",
		Aliases: []string{"ctmpl"},
	}

	cmd.AddCommand(newCmdComponentTemplateList(f))
	cmd.AddCommand(newCmdComponentTemplateGet(f))
	cmd.AddCommand(newCmdComponentTemplateCreate(f))
	cmd.AddCommand(newCmdComponentTemplateDelete(f))

	return cmd
}
