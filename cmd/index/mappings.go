package index

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIndexMappings(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "mappings <index>",
		Short: "Get or update index mappings",
		Long: `Display or update mappings for an Elasticsearch index.

Without -f, displays the current mappings. With -f, updates the mappings from a JSON/YAML file.

Examples:
  # Get index mappings
  es index mappings my-index

  # Update mappings from a file
  es index mappings my-index -f mappings.json

  # Update mappings from stdin
  cat mappings.json | es index mappings my-index -f -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			name := args[0]

			if file != "" {
				var body interface{}
				if err := cmdutil.UnmarshalInput(file, &body); err != nil {
					return err
				}
				if err := c.PutIndexMappings(context.Background(), name, body); err != nil {
					return err
				}
				if !f.Quiet {
					fmt.Fprintf(f.IOStreams.Out, "Mappings updated for index %s.\n", name)
				}
				return nil
			}

			result, err := c.GetIndexMappings(context.Background(), name)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	// Mark as mutating; applies when -f is provided.
	cmd.Annotations = map[string]string{"mutates": "true"}

	return cmd
}
