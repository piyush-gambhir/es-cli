package index

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdIndexCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		file        string
		ifNotExists bool
	)

	cmd := &cobra.Command{
		Use:   "create <index>",
		Short: "Create an index",
		Long: `Create a new Elasticsearch index.

Optionally provide settings and mappings via a JSON/YAML file.

Examples:
  # Create an index with default settings
  es index create my-index

  # Create an index with settings from a file
  es index create my-index -f settings.json

  # Create only if it does not already exist
  es index create my-index --if-not-exists`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			name := args[0]

			var body interface{}
			if file != "" {
				if err := cmdutil.UnmarshalInput(file, &body); err != nil {
					return err
				}
			}

			err = c.CreateIndex(context.Background(), name, body)
			if err != nil {
				if ifNotExists && client.IsConflict(err) {
					if !f.Quiet {
						fmt.Fprintf(f.IOStreams.Out, "Index %s already exists.\n", name)
					}
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Index %s created.\n", name)
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)
	cmdutil.AddIfNotExistsFlag(cmd, &ifNotExists)

	return cmd
}
