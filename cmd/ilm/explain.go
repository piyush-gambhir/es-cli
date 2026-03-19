package ilm

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdILMExplain(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain <index>",
		Short: "Explain ILM status for an index",
		Long: `Show the current ILM lifecycle status for an index.

Examples:
  # Explain ILM for an index
  es ilm explain my-index

  # Output as JSON
  es ilm explain my-index -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			result, err := c.ExplainILM(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	return cmd
}
