package cluster

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdClusterPendingTasks(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "pending-tasks",
		Short: "Show pending cluster tasks",
		Long: `Display the list of pending cluster-level changes that have not yet been executed.

Examples:
  # Show pending tasks
  es cluster pending-tasks

  # Output as JSON
  es cluster pending-tasks -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			tasks, err := c.GetPendingTasks(context.Background())
			if err != nil {
				return err
			}

			if len(tasks) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No pending tasks.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, tasks, &output.TableDef{
				Headers: []string{"Order", "Priority", "Source", "Time In Queue"},
				RowFunc: func(item interface{}) []string {
					t := item.(client.PendingTask)
					return []string{
						fmt.Sprintf("%d", t.InsertOrder),
						t.Priority,
						t.Source,
						t.TimeInQueue,
					}
				},
			})
		},
	}
}
