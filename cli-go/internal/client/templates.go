package client

import (
	"context"
	"encoding/json"
)

type CatTemplate struct {
	Name          string `json:"name"`
	IndexPatterns string `json:"index_patterns"`
	Order         string `json:"order"`
	Version       string `json:"version"`
}

func (c *Client) ListIndexTemplates(ctx context.Context) ([]CatTemplate, error) {
	resp, err := c.Get(ctx, "/_cat/templates?format=json")
	if err != nil {
		return nil, err
	}
	var templates []CatTemplate
	if err := resp.JSON(&templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func (c *Client) GetIndexTemplate(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/_index_template/"+name)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CreateIndexTemplate(ctx context.Context, name string, body interface{}) error {
	resp, err := c.Put(ctx, "/_index_template/"+name, body)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DeleteIndexTemplate(ctx context.Context, name string) error {
	resp, err := c.Delete(ctx, "/_index_template/"+name)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) ListComponentTemplates(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/_component_template")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetComponentTemplate(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/_component_template/"+name)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CreateComponentTemplate(ctx context.Context, name string, body interface{}) error {
	resp, err := c.Put(ctx, "/_component_template/"+name, body)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DeleteComponentTemplate(ctx context.Context, name string) error {
	resp, err := c.Delete(ctx, "/_component_template/"+name)
	if err != nil {
		return err
	}
	return resp.Error()
}
