package cmdutil

import (
	"fmt"
	"io"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/client"
)

func FlagErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func PrintError(w io.Writer, err error) {
	if apiErr, ok := err.(*client.APIError); ok {
		fmt.Fprintf(w, "Error: %s (HTTP %d)\n", apiErr.Message, apiErr.StatusCode)
		return
	}
	fmt.Fprintf(w, "Error: %s\n", err)
}

func ConfirmAction(in io.Reader, out io.Writer, message string, confirmed bool, noInput ...bool) (bool, error) {
	if confirmed {
		return true, nil
	}

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
