package index

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/cmd/index/alias"
	"github.com/piyush-gambhir/es-cli/cmd/index/component_template"
	"github.com/piyush-gambhir/es-cli/cmd/index/template"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdIndex returns the index parent command.
func NewCmdIndex(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "index",
		Short:   "Manage indices",
		Long:    "Create, delete, open, close, and manage Elasticsearch indices, aliases, and templates.",
		Aliases: []string{"indices", "idx"},
	}

	cmd.AddCommand(newCmdIndexList(f))
	cmd.AddCommand(newCmdIndexCreate(f))
	cmd.AddCommand(newCmdIndexGet(f))
	cmd.AddCommand(newCmdIndexDelete(f))
	cmd.AddCommand(newCmdIndexOpen(f))
	cmd.AddCommand(newCmdIndexClose(f))
	cmd.AddCommand(newCmdIndexSettings(f))
	cmd.AddCommand(newCmdIndexMappings(f))
	cmd.AddCommand(newCmdIndexStats(f))
	cmd.AddCommand(newCmdIndexRollover(f))
	cmd.AddCommand(newCmdReindex(f))
	cmd.AddCommand(alias.NewCmdAlias(f))
	cmd.AddCommand(template.NewCmdTemplate(f))
	cmd.AddCommand(componenttemplate.NewCmdComponentTemplate(f))

	return cmd
}
