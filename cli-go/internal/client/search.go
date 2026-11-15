package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

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

func (c *Client) FieldCaps(ctx context.Context, index string, fields string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/%s/_field_caps?%s", index, url.Values{"fields": {fields}}.Encode()))
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}
