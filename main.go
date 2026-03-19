package main

import (
	"fmt"
	"os"

	"github.com/piyush-gambhir/es-cli/cmd"
	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/output"
)

func main() {
	if err := cmd.Execute(); err != nil {
		statusCode := 0
		if apiErr, ok := err.(*client.APIError); ok {
			statusCode = apiErr.StatusCode
		}

		if cmd.OutputFormat == "json" {
			output.WriteError(os.Stderr, "json", err, statusCode)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
