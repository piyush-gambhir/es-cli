package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

func newLoginCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Interactively log in to an Elasticsearch cluster and save a profile",
		Long: `Interactively configure and test a connection to an Elasticsearch cluster.

Prompts for the cluster URL, authentication method (basic, api-key, or bearer),
credentials, and TLS settings. Tests the connection, then saves the configuration
as a named profile.

Examples:
  # Start interactive login
  es login`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.NoInput {
				return fmt.Errorf("interactive input required but --no-input is set. Use environment variables (ES_URL, ES_USERNAME, ES_PASSWORD) instead of 'es login'.")
			}

			reader := bufio.NewReader(os.Stdin)
			out := f.IOStreams.Out

			// Prompt for URL.
			fmt.Fprint(out, "Elasticsearch URL: ")
			urlStr, _ := reader.ReadString('\n')
			urlStr = strings.TrimSpace(urlStr)
			if urlStr == "" {
				return fmt.Errorf("URL is required")
			}

			// Prompt for auth method.
			fmt.Fprint(out, "Auth method (basic/api-key/bearer) [basic]: ")
			authMethod, _ := reader.ReadString('\n')
			authMethod = strings.TrimSpace(authMethod)
			if authMethod == "" {
				authMethod = "basic"
			}

			profile := config.Profile{URL: urlStr, AuthMethod: authMethod}

			switch authMethod {
			case "basic":
				fmt.Fprint(out, "Username: ")
				username, _ := reader.ReadString('\n')
				username = strings.TrimSpace(username)
				password, err := readLoginSecret(reader, out, "Password: ")
				if err != nil {
					return err
				}
				if username == "" || password == "" {
					return fmt.Errorf("username and password are required for basic auth")
				}
				profile.Username = username
				profile.Password = password
			case "api-key":
				fmt.Fprint(out, "API Key ID: ")
				apiKeyID, _ := reader.ReadString('\n')
				apiKeyID = strings.TrimSpace(apiKeyID)
				apiKey, err := readLoginSecret(reader, out, "API Key: ")
				if err != nil {
					return err
				}
				if apiKeyID == "" || apiKey == "" {
					return fmt.Errorf("API key ID and API key are required")
				}
				profile.APIKeyID = apiKeyID
				profile.APIKey = apiKey
				profile.AuthMethod = "api_key"
			case "bearer":
				token, err := readLoginSecret(reader, out, "Bearer Token: ")
				if err != nil {
					return err
				}
				if token == "" {
					return fmt.Errorf("bearer token is required")
				}
				profile.Token = token
			default:
				return fmt.Errorf("invalid auth method: %s (use basic, api-key, or bearer)", authMethod)
			}

			// Prompt for TLS options.
			fmt.Fprint(out, "CA certificate path (leave empty for default): ")
			caCert, _ := reader.ReadString('\n')
			caCert = strings.TrimSpace(caCert)
			if caCert != "" {
				profile.CACert = caCert
			}

			fmt.Fprint(out, "Skip TLS verification? (y/N) [N]: ")
			insecureStr, _ := reader.ReadString('\n')
			insecureStr = strings.TrimSpace(strings.ToLower(insecureStr))
			if insecureStr == "y" || insecureStr == "yes" {
				profile.Insecure = true
			}

			// Test the connection.
			fmt.Fprintln(out, "Testing connection...")
			resolved := &config.ResolvedConfig{
				URL:        profile.URL,
				AuthMethod: profile.AuthMethod,
				Username:   profile.Username,
				Password:   profile.Password,
				APIKeyID:   profile.APIKeyID,
				APIKey:     profile.APIKey,
				Token:      profile.Token,
				CACert:     profile.CACert,
				Insecure:   profile.Insecure,
			}
			c, err := client.NewClient(resolved)
			if err != nil {
				return fmt.Errorf("creating client: %w", err)
			}

			info, err := c.GetClusterInfo(cmd.Context())
			if err != nil {
				return fmt.Errorf("connection test failed: %w", err)
			}

			fmt.Fprintf(out, "Connection successful! Cluster: %s (version %s)\n", info.ClusterName, info.Version.Number)

			// Prompt for profile name.
			fmt.Fprint(out, "Profile name [default]: ")
			profileName, _ := reader.ReadString('\n')
			profileName = strings.TrimSpace(profileName)
			if profileName == "" {
				profileName = "default"
			}

			if err := config.Update(func(cfg *config.Config) error {
				cfg.Profiles[profileName] = profile
				cfg.CurrentProfile = profileName
				return nil
			}); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintf(out, "Profile %q saved and set as current.\n", profileName)
			return nil
		},
	}
}

func readLoginSecret(reader *bufio.Reader, out io.Writer, label string) (string, error) {
	fmt.Fprint(out, label)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		secret, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(out)
		return strings.TrimSpace(string(secret)), err
	}
	secret, err := reader.ReadString('\n')
	return strings.TrimSpace(secret), err
}
