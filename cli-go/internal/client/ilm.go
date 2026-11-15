package client

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) ListILMPolicies(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/_ilm/policy")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetILMPolicy(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/_ilm/policy/%s", name))
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CreateILMPolicy(ctx context.Context, name string, body interface{}) error {
	resp, err := c.Put(ctx, fmt.Sprintf("/_ilm/policy/%s", name), body)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DeleteILMPolicy(ctx context.Context, name string) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/_ilm/policy/%s", name))
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) ExplainILM(ctx context.Context, index string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/%s/_ilm/explain", index))
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}
