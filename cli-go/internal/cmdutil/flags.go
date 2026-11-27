package cmdutil

import "github.com/spf13/cobra"

func AddOutputFlag(cmd *cobra.Command, output *string) {
	cmd.Flags().StringVarP(output, "output", "o", "", "Output format: table, json, yaml")
}

func AddFileFlag(cmd *cobra.Command, file *string) {
	cmd.Flags().StringVarP(file, "file", "f", "", "Path to JSON or YAML file (use - for stdin)")
}

func AddConfirmFlag(cmd *cobra.Command, confirm *bool) {
	cmd.Flags().BoolVar(confirm, "confirm", false, "Skip confirmation prompt")
}

func AddPaginationFlags(cmd *cobra.Command, page, limit *int) {
	cmd.Flags().IntVar(page, "page", 1, "Page number")
	cmd.Flags().IntVar(limit, "limit", 100, "Number of results per page")
}

func AddIfNotExistsFlag(cmd *cobra.Command, ifNotExists *bool) {
	cmd.Flags().BoolVar(ifNotExists, "if-not-exists", false, "Skip creation if the resource already exists (prints a notice on stderr)")
}

func AddIfExistsFlag(cmd *cobra.Command, ifExists *bool) {
	cmd.Flags().BoolVar(ifExists, "if-exists", false, "Succeed silently if resource does not exist (404)")
}
