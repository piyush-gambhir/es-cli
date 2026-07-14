package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for es CLI.

To load completions:

Bash:
  $ source <(es completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ es completion bash > /etc/bash_completion.d/es
  # macOS:
  $ es completion bash > $(brew --prefix)/etc/bash_completion.d/es

Zsh:
  $ source <(es completion zsh)
  # To load completions for each session, execute once:
  $ es completion zsh > "${fpath[1]}/_es"

Fish:
  $ es completion fish | source
  # To load completions for each session, execute once:
  $ es completion fish > ~/.config/fish/completions/es.fish

PowerShell:
  PS> es completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, execute once:
  PS> es completion powershell > es.ps1
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
