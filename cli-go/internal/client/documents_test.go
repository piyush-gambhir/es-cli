package client

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

func TestBulkIndexReturnsPartialFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"took":1,"errors":true,"items":[{"index":{"status":201}},{"index":{"status":400,"error":{"type":"mapper_parsing_exception"}}}]}`))
	}))
	defer ts.Close()

	c, err := NewClient(&config.ResolvedConfig{URL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	c.HTTPClient = ts.Client()
	result, err := c.BulkIndex(context.Background(), "", bytes.NewBufferString("{}\n"))
	if err == nil || !strings.Contains(err.Error(), "1 failed item") {
		t.Fatalf("expected one failed item, got result=%s err=%v", result, err)
	}
	if !bytes.Contains(result, []byte(`"errors":true`)) {
		t.Fatalf("response should be preserved for output: %s", result)
	}
}
