package search

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdSearchQuery(f *cmdutil.Factory) *cobra.Command {
	var (
		file string
		size int
		from int
		sort string
	)

	cmd := &cobra.Command{
		Use:   "query <index>",
		Short: "Run a Query DSL search",
		Long: `Execute a Query DSL search against an Elasticsearch index.

Provide the query body via a JSON/YAML file with -f. Additional parameters
like --size, --from, and --sort are merged into the request body.

Examples:
  # Search with a query file
  es search query my-index -f query.json

  # Search with size and from
  es search query my-index -f query.json --size 20 --from 40

  # Search with sorting
  es search query my-index -f query.json --sort timestamp:desc`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			var body map[string]interface{}
			if file != "" {
				if err := cmdutil.UnmarshalInput(file, &body); err != nil {
					return err
				}
			} else {
				body = make(map[string]interface{})
			}

			if cmd.Flags().Changed("size") {
				body["size"] = size
			}
			if cmd.Flags().Changed("from") {
				body["from"] = from
			}
			if sort != "" {
				body["sort"] = sort
			}

			result, err := c.Search(context.Background(), args[0], body)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)
	cmd.Flags().IntVar(&size, "size", 10, "Number of hits to return")
	cmd.Flags().IntVar(&from, "from", 0, "Starting document offset")
	cmd.Flags().StringVar(&sort, "sort", "", "Sort expression (e.g. timestamp:desc)")

	return cmd
}
