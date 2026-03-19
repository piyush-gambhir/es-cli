package cmdutil

import "github.com/spf13/cobra"

// AddOutputFlag adds the --output/-o flag.
func AddOutputFlag(cmd *cobra.Command, output *string) {
	cmd.Flags().StringVarP(output, "output", "o", "", "Output format: table, json, yaml")
}

// AddFileFlag adds the --file/-f flag.
func AddFileFlag(cmd *cobra.Command, file *string) {
	cmd.Flags().StringVarP(file, "file", "f", "", "Path to JSON or YAML file (use - for stdin)")
}

// AddConfirmFlag adds the --confirm flag.
func AddConfirmFlag(cmd *cobra.Command, confirm *bool) {
	cmd.Flags().BoolVar(confirm, "confirm", false, "Skip confirmation prompt")
}

// AddPaginationFlags adds --page and --limit flags.
func AddPaginationFlags(cmd *cobra.Command, page, limit *int) {
	cmd.Flags().IntVar(page, "page", 1, "Page number")
	cmd.Flags().IntVar(limit, "limit", 100, "Number of results per page")
}

// AddIfNotExistsFlag adds --if-not-exists flag for create commands.
func AddIfNotExistsFlag(cmd *cobra.Command, ifNotExists *bool) {
	cmd.Flags().BoolVar(ifNotExists, "if-not-exists", false, "Succeed silently if resource already exists (409 conflict)")
}

// AddIfExistsFlag adds --if-exists flag for delete commands.
func AddIfExistsFlag(cmd *cobra.Command, ifExists *bool) {
	cmd.Flags().BoolVar(ifExists, "if-exists", false, "Succeed silently if resource does not exist (404)")
}
