package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piyush-gambhir/es-cli/cli-go/internal/config"
)

func TestHotThreadsReturnsHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"type":"security_exception","reason":"action [cluster:monitor/nodes/hot_threads] is unauthorized"},"status":403}`))
	}))
	defer ts.Close()
	c, err := NewClient(&config.ResolvedConfig{URL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.GetHotThreads(context.Background(), "")
	if !IsForbidden(err) || out != "" {
		t.Fatalf("got output %q, err %v; want 403 APIError", out, err)
	}
}
