package search

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdSearchSQL(f *cmdutil.Factory) *cobra.Command {
	var (
		file  string
		query string
	)

	cmd := &cobra.Command{
		Use:   "sql",
		Short: "Run an SQL query",
		Long: `Execute an SQL query against Elasticsearch.

Provide the SQL query inline with --query or from a file with -f.

Examples:
  # Inline SQL query
  es search sql --query "SELECT * FROM my-index LIMIT 10"

  # SQL query from file
  es search sql -f query.sql`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" && file == "" {
				return cmdutil.FlagErrorf("provide a query with --query or a file with -f")
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			sqlQuery := query
			if file != "" {
				data, err := cmdutil.ReadInput(file)
				if err != nil {
					return err
				}
				sqlQuery = string(data)
			}

			result, err := c.SQLQuery(context.Background(), sqlQuery)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)
	cmd.Flags().StringVar(&query, "query", "", "Inline SQL query string")

	return cmd
}
