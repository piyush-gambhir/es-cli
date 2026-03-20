package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/piyush-gambhir/es-cli/internal/build"
	"github.com/piyush-gambhir/es-cli/internal/config"
)

// Client is the Elasticsearch HTTP API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthMethod string // basic, api_key, bearer
	Username   string
	Password   string
	APIKeyID   string
	APIKey     string
	Token      string
	UserAgent  string
}

// NewClient creates a new client from a ResolvedConfig.
func NewClient(rc *config.ResolvedConfig) (*Client, error) {
	if rc.URL == "" {
		return nil, fmt.Errorf("elasticsearch URL is required (use --url, ES_URL, or configure a profile)")
	}

	baseURL := strings.TrimRight(rc.URL, "/")

	tlsConfig, err := buildTLSConfig(rc.CACert, rc.Insecure)
	if err != nil {
		return nil, fmt.Errorf("configuring TLS: %w", err)
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
		TLSClientConfig:     tlsConfig,
	}

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		AuthMethod: rc.AuthMethod,
		Username:   rc.Username,
		Password:   rc.Password,
		APIKeyID:   rc.APIKeyID,
		APIKey:     rc.APIKey,
		Token:      rc.Token,
		UserAgent:  "es-cli/" + build.Version,
	}, nil
}

// Get sends a GET request.
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// Post sends a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

// Put sends a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

// Delete sends a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

// PostRaw sends a POST request with a raw body and custom content type.
// Used for NDJSON bulk requests and other non-JSON payloads.
func (c *Client) PostRaw(ctx context.Context, path string, body io.Reader, contentType string) (*Response, error) {
	url := c.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", contentType)
	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	return &Response{HTTPResponse: resp, RequestURL: url}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*Response, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	return &Response{HTTPResponse: resp, RequestURL: url}, nil
}

// setAuth sets the appropriate authentication header on the request.
func (c *Client) setAuth(req *http.Request) {
	switch c.AuthMethod {
	case "basic":
		req.SetBasicAuth(c.Username, c.Password)
	case "api_key":
		if c.APIKeyID != "" {
			// ID + secret provided separately — encode them
			encoded := base64.StdEncoding.EncodeToString([]byte(c.APIKeyID + ":" + c.APIKey))
			req.Header.Set("Authorization", "ApiKey "+encoded)
		} else {
			// Pre-encoded API key (just the secret, already base64)
			req.Header.Set("Authorization", "ApiKey "+c.APIKey)
		}
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+c.Token)
	default:
		// Auto-detect from available credentials.
		if c.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.Token)
		} else if c.APIKeyID != "" && c.APIKey != "" {
			encoded := base64.StdEncoding.EncodeToString([]byte(c.APIKeyID + ":" + c.APIKey))
			req.Header.Set("Authorization", "ApiKey "+encoded)
		} else if c.APIKey != "" {
			// Pre-encoded API key without ID
			req.Header.Set("Authorization", "ApiKey "+c.APIKey)
		} else if c.Username != "" && c.Password != "" {
			req.SetBasicAuth(c.Username, c.Password)
		}
	}
}
