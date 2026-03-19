package ilm

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
)

// NewCmdILM returns the ILM parent command.
func NewCmdILM(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ilm",
		Short: "Manage Index Lifecycle Management policies",
		Long:  "List, get, create, delete, and explain ILM policies.",
	}

	cmd.AddCommand(newCmdILMList(f))
	cmd.AddCommand(newCmdILMGet(f))
	cmd.AddCommand(newCmdILMCreate(f))
	cmd.AddCommand(newCmdILMDelete(f))
	cmd.AddCommand(newCmdILMExplain(f))

	return cmd
}
