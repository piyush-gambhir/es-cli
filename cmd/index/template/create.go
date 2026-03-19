package template

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdTemplateCreate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an index template",
		Long: `Create or update an index template from a JSON/YAML file.

Examples:
  # Create a template from a file
  es index template create my-template -f template.json

  # Create from stdin
  cat template.json | es index template create my-template -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return cmdutil.FlagErrorf("required flag -f not set")
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			name := args[0]

			var body interface{}
			if err := cmdutil.UnmarshalInput(file, &body); err != nil {
				return err
			}

			if err := c.CreateIndexTemplate(context.Background(), name, body); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Index template %s created.\n", name)
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
