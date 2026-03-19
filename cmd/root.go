package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	cmdconfig "github.com/piyush-gambhir/es-cli/cmd/config"
	"github.com/piyush-gambhir/es-cli/cmd/cluster"
	"github.com/piyush-gambhir/es-cli/cmd/document"
	"github.com/piyush-gambhir/es-cli/cmd/ilm"
	"github.com/piyush-gambhir/es-cli/cmd/index"
	"github.com/piyush-gambhir/es-cli/cmd/ingest"
	"github.com/piyush-gambhir/es-cli/cmd/node"
	"github.com/piyush-gambhir/es-cli/cmd/search"
	"github.com/piyush-gambhir/es-cli/cmd/shard"
	"github.com/piyush-gambhir/es-cli/internal/build"
	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/config"
	"github.com/piyush-gambhir/es-cli/internal/update"
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

// OutputFormat is set during PersistentPreRunE and exported for use by main.go.
var OutputFormat string

// Execute is the main entry point for the CLI.
func Execute() error {
	return newRootCmd().Execute()
}

// loadAndResolveConfig loads the config file and resolves auth from flags/env/config.
func loadAndResolveConfig(cmd *cobra.Command) (*config.ResolvedConfig, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}

	// Determine which profile to use.
	profileName := flagProfile
	if profileName == "" {
		profileName = cfg.CurrentProfile
	}
	var profile *config.Profile
	if profileName != "" {
		p, ok := cfg.Profiles[profileName]
		if ok {
			profile = &p
		}
	}

	// Determine output format.
	output := flagOutput
	if output == "" {
		output = cfg.Defaults.Output
	}

	// Resolve configuration.
	resolved := config.Resolve(flagURL, flagUsername, flagPassword, flagAPIKeyID, flagAPIKey, flagToken, flagCACert, flagInsecure, profile, cfg.Defaults)
	if output != "" {
		resolved.Output = output
	}

	return resolved, cfg, nil
}

// createClient sets up the HTTP client factory on the factory.
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

// checkPermissions enforces read-only and no-input checks.
func checkPermissions(cmd *cobra.Command, resolved *config.ResolvedConfig) error {
	effectiveReadOnly := resolved.ReadOnly // from env > config
	if cmd.Flags().Changed("read-only") {
		effectiveReadOnly = flagReadOnly
	}
	if effectiveReadOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode.\nTo disable, use --read-only=false or remove read_only from your config profile.", cmd.CommandPath())
	}
	return nil
}

func newRootCmd() *cobra.Command {
	f := &cmdutil.Factory{
		IOStreams: cmdutil.DefaultIOStreams(),
	}

	// Channel-based update check result passing from PersistentPreRun to PersistentPostRun.
	var updateResult chan *update.UpdateInfo

	rootCmd := &cobra.Command{
		Use:   "es",
		Short: "Elasticsearch CLI - manage Elasticsearch from the command line",
		Long:  "A command-line interface for managing Elasticsearch clusters, indices, documents, and more.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check env vars for --no-input, --quiet, --verbose.
			if !flagNoInput && os.Getenv("ES_NO_INPUT") != "" {
				flagNoInput = true
			}
			if !flagQuiet && os.Getenv("ES_QUIET") != "" {
				flagQuiet = true
			}
			if !flagVerbose && os.Getenv("ES_VERBOSE") != "" {
				flagVerbose = true
			}
			f.NoInput = flagNoInput
			f.Quiet = flagQuiet
			f.Verbose = flagVerbose

			// Start background update check for most commands.
			cmdName := cmd.Name()
			skipUpdateCheck := cmdName == "update" || cmdName == "version" || cmdName == "completion" || cmdName == "help"
			if !skipUpdateCheck && build.Version != "dev" && build.Version != "" {
				updateResult = make(chan *update.UpdateInfo, 1)
				go func() {
					info, _ := update.CheckForUpdate(build.Version, updateRepo, config.ConfigDir())
					updateResult <- info
				}()
			}

			// Skip auth setup for commands that don't need it.
			if cmdName == "version" || cmdName == "completion" || cmdName == "help" || cmdName == "update" {
				return nil
			}
			// Also skip for config subcommands.
			if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
				return nil
			}

			resolved, cfg, err := loadAndResolveConfig(cmd)
			if err != nil {
				return err
			}

			// Set exported OutputFormat for use by main.go error handler.
			OutputFormat = resolved.Output

			f.Resolved = resolved

			f.Config = func() (*config.Config, error) {
				return cfg, nil
			}

			createClient(f, resolved)

			return checkPermissions(cmd, resolved)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if updateResult == nil {
				return
			}
			select {
			case info := <-updateResult:
				if info != nil && info.Available {
					update.PrintUpdateNotice(os.Stderr, info)
				}
			case <-time.After(2 * time.Second):
				// Don't block command output waiting for update check.
			}
		},
	}

	// Global persistent flags.
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

	// Register subcommands.
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
