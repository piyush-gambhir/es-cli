package cmdutil

import (
	"io"
	"os"

	"github.com/piyush-gambhir/es-cli/internal/client"
	"github.com/piyush-gambhir/es-cli/internal/config"
)

// Factory provides shared dependencies to all commands.
type Factory struct {
	Config   func() (*config.Config, error)
	Client   func() (*client.Client, error)
	IOStreams IOStreams
	// Resolved holds the resolved config after PersistentPreRunE.
	Resolved *config.ResolvedConfig
	// NoInput disables all interactive prompts (for CI/agent use).
	NoInput bool
	// Quiet suppresses informational output.
	Quiet bool
	// Verbose enables verbose HTTP logging.
	Verbose bool
}

// IOStreams holds standard I/O streams.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// DefaultIOStreams returns IOStreams connected to os.Stdin, os.Stdout, os.Stderr.
func DefaultIOStreams() IOStreams {
	return IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}
