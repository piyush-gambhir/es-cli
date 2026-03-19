package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/piyush-gambhir/es-cli/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/internal/config"
)

// NewCmdConfig returns the config parent command.
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "View and modify the es CLI configuration file.",
	}

	cmd.AddCommand(newCmdConfigView(f))
	cmd.AddCommand(newCmdConfigSet(f))
	cmd.AddCommand(newCmdConfigUseProfile(f))
	cmd.AddCommand(newCmdConfigListProfiles(f))

	return cmd
}

func newCmdConfigView(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Display the current configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshaling config: %w", err)
			}

			fmt.Fprint(f.IOStreams.Out, string(data))
			return nil
		},
	}
}

func newCmdConfigSet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value. Supported keys:
  defaults.output    - Default output format (table, json, yaml)
  current_profile    - Current profile name`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			switch key {
			case "defaults.output":
				switch value {
				case "table", "json", "yaml":
					cfg.Defaults.Output = value
				default:
					return fmt.Errorf("invalid output format: %s (use table, json, or yaml)", value)
				}
			case "current_profile":
				if err := cfg.SetCurrentProfile(value); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Set %s = %s\n", key, value)
			}
			return nil
		},
	}
}

func newCmdConfigUseProfile(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "use-profile <name>",
		Short: "Switch to a different profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := cfg.SetCurrentProfile(args[0]); err != nil {
				return err
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Switched to profile %q\n", args[0])
			}
			return nil
		},
	}
}

func newCmdConfigListProfiles(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list-profiles",
		Short: "List all configured profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if len(cfg.Profiles) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No profiles configured. Run 'es login' to create one.")
				return nil
			}

			for name, profile := range cfg.Profiles {
				marker := "  "
				if name == cfg.CurrentProfile {
					marker = "* "
				}
				authType := profile.AuthMethod
				if authType == "" {
					authType = "auto"
				}
				fmt.Fprintf(f.IOStreams.Out, "%s%s (%s, %s)\n", marker, name, profile.URL, authType)
			}

			return nil
		},
	}
}
