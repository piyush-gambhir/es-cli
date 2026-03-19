package ingest

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdIngest returns the ingest parent command.
func NewCmdIngest(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ingest",
		Aliases: []string{"pipeline"},
		Short:   "Manage ingest pipelines",
		Long:    "List, get, create, delete, and simulate ingest pipelines.",
	}

	cmd.AddCommand(newCmdIngestList(f))
	cmd.AddCommand(newCmdIngestGet(f))
	cmd.AddCommand(newCmdIngestCreate(f))
	cmd.AddCommand(newCmdIngestDelete(f))
	cmd.AddCommand(newCmdIngestSimulate(f))

	return cmd
}
