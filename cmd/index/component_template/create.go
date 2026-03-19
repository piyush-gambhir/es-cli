package componenttemplate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdComponentTemplateCreate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a component template",
		Long: `Create or update a component template from a JSON/YAML file.

Examples:
  # Create a component template from a file
  es index component-template create my-component -f component.json

  # Create from stdin
  cat component.json | es index component-template create my-component -f -`,
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

			if err := c.CreateComponentTemplate(context.Background(), name, body); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Component template %s created.\n", name)
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
