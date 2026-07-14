package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// authHeaders is the set of headers that contain secrets and must be redacted.
var authHeaders = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
	"Set-Cookie":    true,
	"X-Api-Key":     true,
}

// redactAuthHeaders returns a shallow copy of the header map with auth-sensitive
// headers replaced by "[REDACTED]". Non-secret headers are kept intact.
func redactAuthHeaders(h http.Header) http.Header {
	redacted := make(http.Header, len(h))
	for key, values := range h {
		if authHeaders[key] {
			redacted[key] = []string{"[REDACTED]"}
		} else {
			redacted[key] = values
		}
	}
	return redacted
}

// verboseTransport wraps an http.RoundTripper and logs requests/responses to stderr.
type verboseTransport struct {
	inner  http.RoundTripper
	output io.Writer
}

func (t *verboseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Log request.
	fmt.Fprintf(t.output, "--> %s %s\n", req.Method, req.URL.String())

	// Log headers with auth secrets redacted.
	for key, values := range redactAuthHeaders(req.Header) {
		for _, v := range values {
			fmt.Fprintf(t.output, "    %s: %s\n", key, v)
		}
	}

	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(t.output, "<-- ERROR: %v (%s)\n", err, time.Since(start).Round(time.Millisecond))
		return nil, err
	}

	fmt.Fprintf(t.output, "<-- %d %s (%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), time.Since(start).Round(time.Millisecond))

	// Log response headers with auth secrets redacted.
	for key, values := range redactAuthHeaders(resp.Header) {
		for _, v := range values {
			fmt.Fprintf(t.output, "    %s: %s\n", key, v)
		}
	}

	return resp, nil
}

// EnableVerboseLogging wraps the client's HTTP transport with verbose logging.
func (c *Client) EnableVerboseLogging(w io.Writer) {
	transport := c.HTTPClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	c.HTTPClient.Transport = &verboseTransport{inner: transport, output: w}
}

// buildTLSConfig creates a TLS configuration supporting custom CA certificates
// and insecure skip verification.
func buildTLSConfig(caCertPath string, insecure bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
	}

	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("reading CA certificate %s: %w", caCertPath, err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate from %s", caCertPath)
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}
