package document

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdDocument returns the document parent command.
func NewCmdDocument(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "document",
		Aliases: []string{"doc", "docs"},
		Short:   "Manage Elasticsearch documents",
		Long:    "Get, index, delete, bulk index, and multi-get documents.",
	}

	cmd.AddCommand(newCmdDocumentGet(f))
	cmd.AddCommand(newCmdDocumentIndex(f))
	cmd.AddCommand(newCmdDocumentDelete(f))
	cmd.AddCommand(newCmdDocumentBulk(f))
	cmd.AddCommand(newCmdDocumentMGet(f))

	return cmd
}
