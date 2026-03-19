package ingest

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

func newCmdIngestDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		confirm  bool
		ifExists bool
	)

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an ingest pipeline",
		Long: `Delete an ingest pipeline by name.

Requires confirmation unless --confirm is provided.

Examples:
  # Delete a pipeline (with confirmation prompt)
  es ingest delete my-pipeline

  # Delete without confirmation
  es ingest delete my-pipeline --confirm

  # Ignore if pipeline does not exist
  es ingest delete my-pipeline --confirm --if-exists`,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"mutates": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := cmdutil.ConfirmAction(f.IOStreams.In, f.IOStreams.ErrOut,
				fmt.Sprintf("Delete pipeline %s?", args[0]),
				confirm, f.NoInput)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			err = c.DeletePipeline(context.Background(), args[0])
			if err != nil {
				if ifExists && client.IsNotFound(err) {
					return nil
				}
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.ErrOut, "Pipeline %s deleted\n", args[0])
			}

			return nil
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm)
	cmdutil.AddIfExistsFlag(cmd, &ifExists)

	return cmd
}
