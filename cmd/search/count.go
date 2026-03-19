package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdSearchCount(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "count <index>",
		Short: "Count documents in an index",
		Long: `Count the number of documents matching a query in an index.

Optionally provide a query body via -f to count matching documents.

Examples:
  # Count all documents
  es search count my-index

  # Count with a query filter
  es search count my-index -f query.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			var body interface{}
			if file != "" {
				var parsed map[string]interface{}
				if err := cmdutil.UnmarshalInput(file, &parsed); err != nil {
					return err
				}
				body = parsed
			}

			result, err := c.Count(context.Background(), args[0], body)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, &output.TableDef{
				Headers: []string{"Count"},
				RowFunc: func(item interface{}) []string {
					raw := item.(json.RawMessage)
					var countResult struct {
						Count int64 `json:"count"`
					}
					_ = json.Unmarshal(raw, &countResult)
					return []string{fmt.Sprintf("%d", countResult.Count)}
				},
			})
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
