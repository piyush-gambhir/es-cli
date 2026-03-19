package index

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func newCmdIndexSettings(f *cmdutil.Factory) *cobra.Command {
	var set []string

	cmd := &cobra.Command{
		Use:   "settings <index>",
		Short: "Get or update index settings",
		Long: `Display or update settings for an Elasticsearch index.

Without --set, displays the current settings. With --set, updates the specified settings.

Examples:
  # Get index settings
  es index settings my-index

  # Update a setting
  es index settings my-index --set number_of_replicas=2

  # Update multiple settings
  es index settings my-index --set number_of_replicas=2 --set refresh_interval=30s`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			name := args[0]

			if len(set) > 0 {
				settings := make(map[string]interface{})
				for _, s := range set {
					parts := strings.SplitN(s, "=", 2)
					if len(parts) != 2 {
						return cmdutil.FlagErrorf("invalid --set format %q, expected key=value", s)
					}
					settings[parts[0]] = parts[1]
				}
				body := map[string]interface{}{
					"index": settings,
				}
				if err := c.PutIndexSettings(context.Background(), name, body); err != nil {
					return err
				}
				if !f.Quiet {
					fmt.Fprintf(f.IOStreams.Out, "Settings updated for index %s.\n", name)
				}
				return nil
			}

			result, err := c.GetIndexSettings(context.Background(), name)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, result, nil)
		},
	}

	cmd.Flags().StringArrayVar(&set, "set", nil, "Set a setting as key=value (repeatable)")

	// Mark as mutating only when --set is provided; the annotation is checked
	// at runtime by the root PersistentPreRunE, so we set it unconditionally
	// and let the read path run harmlessly.
	cmd.Annotations = map[string]string{"mutates": "true"}

	return cmd
}
