package cmdutil

import (
	"io"
	"os"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

type Factory struct {
	Config    func() (*config.Config, error)
	Client    func() (*client.Client, error)
	IOStreams IOStreams
	Resolved  *config.ResolvedConfig
	NoInput   bool
	Quiet     bool
	Verbose   bool
}

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

func DefaultIOStreams() IOStreams {
	return IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}
