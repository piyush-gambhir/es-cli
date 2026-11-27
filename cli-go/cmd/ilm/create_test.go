package ilm

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

func TestCreateIfNotExistsDoesNotOverwrite(t *testing.T) {
	file := filepath.Join(t.TempDir(), "policy.json")
	if err := os.WriteFile(file, []byte(`{"policy":{"phases":{}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		get     int
		wantPut bool
		wantErr bool
	}{
		{name: "exists", get: http.StatusOK},
		{name: "missing", get: http.StatusNotFound, wantPut: true},
		{name: "unauthorized", get: http.StatusUnauthorized, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			put := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_ilm/policy/p1" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				w.Header().Set("Content-Type", "application/json")
				switch r.Method {
				case http.MethodGet:
					w.WriteHeader(tc.get)
					_, _ = w.Write([]byte(`{}`))
				case http.MethodPut:
					put = true
					_, _ = w.Write([]byte(`{"acknowledged":true}`))
				}
			}))
			defer ts.Close()
			var out bytes.Buffer
			f := &cmdutil.Factory{
				IOStreams: cmdutil.IOStreams{In: &bytes.Buffer{}, Out: &out, ErrOut: &out},
				Client: func() (*client.Client, error) {
					return client.NewClient(&config.ResolvedConfig{URL: ts.URL})
				},
			}
			cmd := newCmdILMCreate(f)
			cmd.SetArgs([]string{"p1", "-f", file, "--if-not-exists"})
			err := cmd.Execute()
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, want error %t", err, tc.wantErr)
			}
			if put != tc.wantPut {
				t.Fatalf("PUT sent = %t, want %t", put, tc.wantPut)
			}
		})
	}
}
