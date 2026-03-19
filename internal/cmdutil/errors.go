package cmdutil

import (
	"fmt"
	"io"

	"github.com/piyush-gambhir/es-cli/internal/client"
)

// FlagErrorf creates a formatted error for flag-related issues.
func FlagErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// PrintError writes a formatted error message to the given writer.
func PrintError(w io.Writer, err error) {
	if apiErr, ok := err.(*client.APIError); ok {
		fmt.Fprintf(w, "Error: %s (HTTP %d)\n", apiErr.Message, apiErr.StatusCode)
		return
	}
	fmt.Fprintf(w, "Error: %s\n", err)
}

// ConfirmAction prompts the user for confirmation if not auto-confirmed.
// If noInput is true and confirmed is false, returns an error indicating
// that interactive input is required.
func ConfirmAction(in io.Reader, out io.Writer, message string, confirmed bool, noInput ...bool) (bool, error) {
	if confirmed {
		return true, nil
	}

	// Check if no-input mode is active (optional parameter for backwards compat).
	if len(noInput) > 0 && noInput[0] {
		return false, fmt.Errorf("interactive input required but --no-input is set. Use --confirm for destructive operations.")
	}

	fmt.Fprintf(out, "%s [y/N]: ", message)
	var response string
	_, err := fmt.Fscan(in, &response)
	if err != nil {
		return false, nil
	}

	return response == "y" || response == "Y" || response == "yes" || response == "Yes", nil
}
