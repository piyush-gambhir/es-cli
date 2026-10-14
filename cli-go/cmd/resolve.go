package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/output"
)

// resolution describes arguments using the parser without executing their command.
type resolution struct {
	Command     string            `json:"command"`
	Effect      string            `json:"effect"`
	Conditional bool              `json:"conditional"`
	Interactive bool              `json:"interactive"`
	Secret      bool              `json:"secret_output"`
	Flags       map[string]string `json:"flags"`
	Args        []string          `json:"args"`
}

func newCommandsCmd() *cobra.Command {
	commands := &cobra.Command{Use: "commands", Short: "Offline command discovery for agents"}
	commands.AddCommand(&cobra.Command{
		Use:                "resolve -- <arguments>",
		Short:              "Show the command, effect and flags that arguments would run, without running them",
		Example:            "  es commands resolve -- cluster health\n  es commands resolve -- -o json index delete my-index --confirm",
		DisableFlagParsing: true,
		// Override the root hook to avoid config, credentials and update checks.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			OutputFormat = "json"
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for i, arg := range args {
				if arg == "--" {
					args = args[i+1:]
					break
				}
			}
			probe := newRootCmd()
			target, rest, err := probe.Find(args)
			if err != nil || target == probe || !target.Runnable() {
				return fmt.Errorf("arguments do not name a runnable es command")
			}
			if err := target.ParseFlags(rest); err != nil {
				return err
			}
			return output.Print(cmd.OutOrStdout(), "json", resolveCommand(target), nil)
		},
	})
	return commands
}

func resolveCommand(target *cobra.Command) resolution {
	res := resolution{
		Command: target.CommandPath(), Effect: "read",
		Flags: map[string]string{}, Args: append([]string{}, target.Flags().Args()...),
	}
	ann := target.Annotations
	flagSet := func(name string) bool { return name != "" && target.Flags().Changed(name) }
	switch {
	// Checked first: these commands also carry mutates=true but write only with the flag.
	case ann["mutates-when-flag"] != "":
		if flagSet(ann["mutates-when-flag"]) {
			res.Effect = "remote_write"
		}
	case ann["mutates"] == "true" || ann["mutates"] == "conditional":
		res.Effect, res.Conditional = "remote_write", ann["mutates"] == "conditional"
	case ann["writes-local"] == "true" || ann["writes-local"] == "conditional" || ann["self-update"] != "" || flagSet(ann["writes-local-when-flag"]):
		res.Effect, res.Conditional = "local_write", ann["writes-local"] == "conditional"
	}
	res.Interactive = ann["interactive"] == "true"
	res.Secret = ann["secret-output"] == "true"
	target.Flags().Visit(func(f *pflag.Flag) { res.Flags[f.Name] = f.Value.String() })
	return res
}
