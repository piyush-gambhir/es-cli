package cluster

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

func TestPendingTasksEmptyJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tasks":[]}`))
	}))
	defer ts.Close()
	var out bytes.Buffer
	f := &cmdutil.Factory{
		IOStreams: cmdutil.IOStreams{In: &bytes.Buffer{}, Out: &out, ErrOut: &bytes.Buffer{}},
		Resolved:  &config.ResolvedConfig{Output: "json"},
		Client: func() (*client.Client, error) {
			return client.NewClient(&config.ResolvedConfig{URL: ts.URL})
		},
	}
	cmd := newCmdClusterPendingTasks(f)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "[]" {
		t.Fatalf("output = %q, want []", got)
	}
}
