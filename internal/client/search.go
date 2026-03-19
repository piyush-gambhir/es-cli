package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Search executes a search query against the given index.
func (c *Client) Search(ctx context.Context, index string, body interface{}) (json.RawMessage, error) {
	resp, err := c.Post(ctx, fmt.Sprintf("/%s/_search", index), body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// SQLQuery executes an SQL query against Elasticsearch.
func (c *Client) SQLQuery(ctx context.Context, query string) (json.RawMessage, error) {
	body := map[string]string{"query": query}
	resp, err := c.Post(ctx, "/_sql", body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Count returns the number of documents matching a query.
func (c *Client) Count(ctx context.Context, index string, body interface{}) (json.RawMessage, error) {
	resp, err := c.Post(ctx, fmt.Sprintf("/%s/_count", index), body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// MultiSearch executes a multi-search request with NDJSON body.
func (c *Client) MultiSearch(ctx context.Context, body io.Reader) (json.RawMessage, error) {
	resp, err := c.PostRaw(ctx, "/_msearch", body, "application/x-ndjson")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// FieldCaps returns field capabilities for the given index and fields.
func (c *Client) FieldCaps(ctx context.Context, index string, fields string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/%s/_field_caps?fields=%s", index, fields))
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}
