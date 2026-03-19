package ilm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdILMCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		file        string
		ifNotExists bool
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an ILM policy",
		Long: `Create a new ILM policy from a JSON/YAML definition file.

Examples:
  # Create an ILM policy
  es ilm create my-policy -f policy.json

  # Create only if it doesn't exist
  es ilm create my-policy -f policy.json --if-not-exists`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"mutates": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return cmdutil.FlagErrorf("file is required (-f)")
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			var body interface{}
			if err := cmdutil.UnmarshalInput(file, &body); err != nil {
				return err
			}

			err = c.CreateILMPolicy(context.Background(), args[0], body)
			if err != nil {
				if ifNotExists {
					// Policy already exists; treat as success.
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.ErrOut, "ILM policy %s created\n", args[0])
			}

			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)
	cmdutil.AddIfNotExistsFlag(cmd, &ifNotExists)

	return cmd
}
