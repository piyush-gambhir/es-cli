package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// GetDocument retrieves a single document by index and ID.
func (c *Client) GetDocument(ctx context.Context, index, id string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/%s/_doc/%s", index, id))
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// IndexDocument indexes a document. If id is empty, a POST is used to auto-generate an ID.
// If id is provided, a PUT is used to index with the given ID.
func (c *Client) IndexDocument(ctx context.Context, index string, id string, body interface{}) (json.RawMessage, error) {
	var resp *Response
	var err error
	if id == "" {
		resp, err = c.Post(ctx, fmt.Sprintf("/%s/_doc", index), body)
	} else {
		resp, err = c.Put(ctx, fmt.Sprintf("/%s/_doc/%s", index, id), body)
	}
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteDocument deletes a document by index and ID.
func (c *Client) DeleteDocument(ctx context.Context, index, id string) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/%s/_doc/%s", index, id))
	if err != nil {
		return err
	}
	return resp.Error()
}

// BulkIndex sends a bulk indexing request with NDJSON body.
// If index is provided, the path is /<index>/_bulk, otherwise /_bulk.
func (c *Client) BulkIndex(ctx context.Context, index string, body io.Reader) (json.RawMessage, error) {
	path := "/_bulk"
	if index != "" {
		path = fmt.Sprintf("/%s/_bulk", index)
	}
	resp, err := c.PostRaw(ctx, path, body, "application/x-ndjson")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// MultiGet retrieves multiple documents in a single request.
func (c *Client) MultiGet(ctx context.Context, index string, body interface{}) (json.RawMessage, error) {
	resp, err := c.Post(ctx, fmt.Sprintf("/%s/_mget", index), body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}
