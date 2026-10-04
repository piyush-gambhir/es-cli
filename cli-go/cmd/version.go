package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/update"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long: `Print the es-cli version, commit hash, and build date.

When an earlier update check cached the latest release, also print it and
whether an update is available. version itself never contacts GitHub; run
es update --check for a fresh answer.

Examples:
  es version`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "es-cli version %s\n", build.Version)
			fmt.Fprintf(w, "  commit: %s\n", build.Commit)
			fmt.Fprintf(w, "  built:  %s\n", build.Date)
			if !update.IsReleaseVersion(build.Version) {
				return
			}
			if info, ok := update.Cached(build.Version, config.ConfigDir()); ok {
				fmt.Fprintf(w, "  latest: %s\n", info.LatestVersion)
				fmt.Fprintf(w, "  update_available: %t\n", info.Available)
			}
		},
	}
}
