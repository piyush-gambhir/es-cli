package index

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

func TestCreateIfNotExistsAcceptsExistingIndex(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"type":"resource_already_exists_exception","reason":"index [logs/abc] already exists"},"status":400}`))
	}))
	defer ts.Close()
	for _, tc := range []struct {
		args    []string
		wantErr bool
	}{
		{args: []string{"logs", "--if-not-exists"}},
		{args: []string{"logs"}, wantErr: true},
	} {
		var out bytes.Buffer
		f := &cmdutil.Factory{
			IOStreams: cmdutil.IOStreams{In: &bytes.Buffer{}, Out: &out, ErrOut: &out},
			Client: func() (*client.Client, error) {
				return client.NewClient(&config.ResolvedConfig{URL: ts.URL})
			},
		}
		cmd := newCmdIndexCreate(f)
		cmd.SetArgs(tc.args)
		if err := cmd.Execute(); (err != nil) != tc.wantErr {
			t.Fatalf("%v: error = %v, want error %t", tc.args, err, tc.wantErr)
		}
	}
}
