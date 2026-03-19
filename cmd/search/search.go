package search

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdSearch returns the search parent command.
func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Aliases: []string{"query"},
		Short:   "Search and query Elasticsearch",
		Long:    "Run search queries, SQL queries, count documents, multi-search, and field capabilities.",
	}

	cmd.AddCommand(newCmdSearchQuery(f))
	cmd.AddCommand(newCmdSearchSQL(f))
	cmd.AddCommand(newCmdSearchCount(f))
	cmd.AddCommand(newCmdSearchMSearch(f))
	cmd.AddCommand(newCmdSearchFieldCaps(f))

	return cmd
}
