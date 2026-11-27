package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

func TestListIndicesFiltersByStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("expand_wildcards"); got != "closed" {
			t.Errorf("expand_wildcards = %q, want closed", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"index":"old","status":"close"},{"index":".ds-logs-2026.11.27-000001","status":"open"}]`))
	}))
	defer ts.Close()
	c, err := NewClient(&config.ResolvedConfig{URL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	indices, err := c.ListIndices(context.Background(), "", "", "close")
	if err != nil {
		t.Fatal(err)
	}
	if len(indices) != 1 || indices[0].Index != "old" {
		t.Fatalf("got %+v, want only the closed index", indices)
	}
}
