package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/es-cli/cli-go/cmd/cluster"
	cmdconfig "github.com/piyush-gambhir/es-cli/cli-go/cmd/config"
	"github.com/piyush-gambhir/es-cli/cli-go/cmd/document"
	"github.com/piyush-gambhir/es-cli/cli-go/cmd/ilm"
	"github.com/piyush-gambhir/es-cli/cli-go/cmd/index"
	"github.com/piyush-gambhir/es-cli/cli-go/cmd/ingest"
	"github.com/piyush-gambhir/es-cli/cli-go/cmd/node"
	"github.com/piyush-gambhir/es-cli/cli-go/cmd/search"
	"github.com/piyush-gambhir/es-cli/cli-go/cmd/shard"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/update"
)

var (
	flagOutput   string
	flagProfile  string
	flagURL      string
	flagUsername string
	flagPassword string
	flagAPIKeyID string
	flagAPIKey   string
	flagToken    string
	flagCACert   string
	flagInsecure bool
	flagReadOnly bool
	flagNoInput  bool
	flagQuiet    bool
	flagVerbose  bool
)

var OutputFormat string

var (
	stderrIsTerminal = func() bool { return term.IsTerminal(int(os.Stderr.Fd())) }
	startUpdateCheck = update.StartCheck
)

func Execute() error {
	if runtime.GOOS == "windows" {
		if exe, err := update.ExecutablePath(); err == nil {
			update.RemoveOldExecutable(exe)
		}
	}
	return newRootCmd().Execute()
}

func commandNames(cmd *cobra.Command) []string {
	var names []string
	for c := cmd; c != nil && c.HasParent(); c = c.Parent() {
		names = append(names, c.Name())
	}
	return names
}

func loadAndResolveConfig(cmd *cobra.Command) (*config.ResolvedConfig, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}

	profileName := flagProfile
	if profileName == "" {
		profileName = cfg.CurrentProfile
	}
	var profile *config.Profile
	if profileName != "" {
		p, ok := cfg.Profiles[profileName]
		if !ok && cmd.Name() != "login" {
			return nil, nil, fmt.Errorf("profile %q not found", profileName)
		}
		if ok {
			profile = &p
		}
	}

	output := flagOutput
	if output == "" {
		output = cfg.Defaults.Output
	}

	resolved := config.Resolve(flagURL, flagUsername, flagPassword, flagAPIKeyID, flagAPIKey, flagToken, flagCACert, flagInsecure, profile, cfg.Defaults)
	if cmd.Flags().Changed("insecure") {
		resolved.Insecure = flagInsecure
	}
	if output != "" {
		resolved.Output = output
	}

	return resolved, cfg, nil
}

func envFlagEnabled(name string) bool {
	v := strings.TrimSpace(os.Getenv(name))
	return strings.EqualFold(v, "true") || v == "1"
}

func createClient(f *cmdutil.Factory, resolved *config.ResolvedConfig) {
	f.Client = func() (*client.Client, error) {
		c, err := client.NewClient(resolved)
		if err != nil {
			return nil, err
		}
		if flagVerbose {
			c.EnableVerboseLogging(f.IOStreams.ErrOut)
		}
		return c, nil
	}
}

func checkPermissions(cmd *cobra.Command, resolved *config.ResolvedConfig) error {
	effectiveReadOnly := resolved.ReadOnly
	if flagReadOnly {
		effectiveReadOnly = true
	}
	if effectiveReadOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		if flag := cmd.Annotations["mutates-when-flag"]; flag != "" && !cmd.Flags().Changed(flag) {
			return nil
		}
		return fmt.Errorf("command '%s' is blocked in read-only mode; remove read_only from the profile or disable the read-only environment setting to permit writes", cmd.CommandPath())
	}
	return nil
}

func newRootCmd() *cobra.Command {
	f := &cmdutil.Factory{
		IOStreams: cmdutil.DefaultIOStreams(),
	}

	var updateCheck *update.Check

	rootCmd := &cobra.Command{
		Use:   "es",
		Short: "Elasticsearch CLI - manage Elasticsearch from the command line",
		Long: `A command-line interface for managing Elasticsearch clusters, indices, documents, and more.

Full command reference (for agents/LLMs): https://es-cli.pages.dev/llms.txt
Claude Code skill: https://github.com/piyush-gambhir/es-cli/blob/main/SKILL.md`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if envFlagEnabled("ES_NO_INPUT") {
				flagNoInput = true
			}
			if !cmd.Flags().Changed("quiet") {
				flagQuiet = envFlagEnabled("ES_QUIET")
			}
			if !cmd.Flags().Changed("verbose") {
				flagVerbose = envFlagEnabled("ES_VERBOSE")
			}
			f.NoInput = flagNoInput
			f.Quiet = flagQuiet
			f.Verbose = flagVerbose

			if update.ShouldCheck(update.NotifierOptions{
				Version:          build.Version,
				Commands:         commandNames(cmd),
				Quiet:            flagQuiet,
				StderrIsTerminal: stderrIsTerminal(),
				Getenv:           os.Getenv,
			}) {
				updateCheck = startUpdateCheck(build.Version, config.ConfigDir())
			}

			cmdName := cmd.Name()
			if cmdName == "version" || cmdName == "completion" || cmdName == "help" || cmdName == "update" {
				return nil
			}
			if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
				return nil
			}

			resolved, cfg, err := loadAndResolveConfig(cmd)
			if err != nil {
				return err
			}

			OutputFormat = resolved.Output

			f.Resolved = resolved

			f.Config = func() (*config.Config, error) {
				return cfg, nil
			}

			createClient(f, resolved)

			return checkPermissions(cmd, resolved)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if updateCheck == nil {
				return
			}
			exe, _ := executablePath()
			updateCheck.Notify(f.IOStreams.ErrOut, installMethodOf(exe))
		},
	}

	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVar(&flagProfile, "profile", "", "Configuration profile to use")
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "Elasticsearch URL")
	rootCmd.PersistentFlags().StringVarP(&flagUsername, "username", "u", "", "Username for basic auth")
	rootCmd.PersistentFlags().StringVarP(&flagPassword, "password", "p", "", "Password for basic auth")
	rootCmd.PersistentFlags().StringVar(&flagAPIKeyID, "api-key-id", "", "API key ID")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key secret")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Bearer token")
	rootCmd.PersistentFlags().StringVar(&flagCACert, "ca-cert", "", "Path to CA certificate for TLS")
	rootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "k", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().BoolVar(&flagReadOnly, "read-only", false, "Block write operations (safety mode for agents)")
	rootCmd.PersistentFlags().BoolVar(&flagNoInput, "no-input", false, "Disable all interactive prompts (for CI/agent use)")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress informational output")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose HTTP logging")

	rootCmd.AddCommand(newCommandsCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(newLoginCmd(f))
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(cmdconfig.NewCmdConfig(f))
	rootCmd.AddCommand(cluster.NewCmdCluster(f))
	rootCmd.AddCommand(node.NewCmdNode(f))
	rootCmd.AddCommand(index.NewCmdIndex(f))
	rootCmd.AddCommand(shard.NewCmdShard(f))
	rootCmd.AddCommand(search.NewCmdSearch(f))
	rootCmd.AddCommand(document.NewCmdDocument(f))
	rootCmd.AddCommand(ingest.NewCmdIngest(f))
	rootCmd.AddCommand(ilm.NewCmdILM(f))

	return rootCmd
}
