package alias

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdAliasCreate(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "create <index> <alias>",
		Short: "Create an alias",
		Long: `Create an alias for an Elasticsearch index.

Examples:
  # Create an alias
  es index alias create my-index my-alias`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			index := args[0]
			alias := args[1]

			if err := c.CreateAlias(context.Background(), index, alias); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Alias %s created for index %s.\n", alias, index)
			}
			return nil
		},
	}
}
