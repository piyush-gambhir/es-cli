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

func TestDocumentIDIsOnePathSegment(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method || r.URL.EscapedPath() != "/logs/_doc/a%2Fb%3Fc%23d%25" || r.URL.RawQuery != "" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer ts.Close()
			c, err := NewClient(&config.ResolvedConfig{URL: ts.URL})
			if err != nil {
				t.Fatal(err)
			}
			switch method {
			case http.MethodGet:
				_, err = c.GetDocument(context.Background(), "logs", "a/b?c#d%")
			case http.MethodPut:
				_, err = c.IndexDocument(context.Background(), "logs", "a/b?c#d%", map[string]string{"value": "test"})
			case http.MethodDelete:
				err = c.DeleteDocument(context.Background(), "logs", "a/b?c#d%")
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFieldCapsEncodesFields(t *testing.T) {
	fields := "message&include_unmapped=true,#percent%"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logs-*,metrics-*/_field_caps" || r.URL.Query().Get("fields") != fields || len(r.URL.Query()) != 1 {
			t.Errorf("unexpected request URL: %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	c, err := NewClient(&config.ResolvedConfig{URL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.FieldCaps(context.Background(), "logs-*,metrics-*", fields); err != nil {
		t.Fatal(err)
	}
}
