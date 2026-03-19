package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// CatIndex represents an index from _cat/indices.
type CatIndex struct {
	Index        string `json:"index"`
	Health       string `json:"health"`
	Status       string `json:"status"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
}

// ListIndices returns indices from _cat/indices.
func (c *Client) ListIndices(ctx context.Context, pattern, health, status string) ([]CatIndex, error) {
	path := "/_cat/indices"
	if pattern != "" {
		path += "/" + pattern
	}
	path += "?format=json&h=index,health,status,uuid,pri,rep,docs.count,docs.deleted,store.size,pri.store.size"
	if health != "" {
		path += "&health=" + health
	}
	if status != "" {
		path += "&expand_wildcards=" + status
	}
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var indices []CatIndex
	if err := resp.JSON(&indices); err != nil {
		return nil, err
	}
	return indices, nil
}

// CreateIndex creates a new index with optional settings/mappings body.
func (c *Client) CreateIndex(ctx context.Context, name string, body interface{}) error {
	resp, err := c.Put(ctx, "/"+name, body)
	if err != nil {
		return err
	}
	return resp.Error()
}

// GetIndex returns the full index definition.
func (c *Client) GetIndex(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/"+name)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteIndex deletes an index.
func (c *Client) DeleteIndex(ctx context.Context, name string) error {
	resp, err := c.Delete(ctx, "/"+name)
	if err != nil {
		return err
	}
	return resp.Error()
}

// OpenIndex opens a closed index.
func (c *Client) OpenIndex(ctx context.Context, name string) error {
	resp, err := c.Post(ctx, "/"+name+"/_open", nil)
	if err != nil {
		return err
	}
	return resp.Error()
}

// CloseIndex closes an open index.
func (c *Client) CloseIndex(ctx context.Context, name string) error {
	resp, err := c.Post(ctx, "/"+name+"/_close", nil)
	if err != nil {
		return err
	}
	return resp.Error()
}

// GetIndexSettings returns settings for an index.
func (c *Client) GetIndexSettings(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/"+name+"/_settings")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// PutIndexSettings updates settings for an index.
func (c *Client) PutIndexSettings(ctx context.Context, name string, body interface{}) error {
	resp, err := c.Put(ctx, "/"+name+"/_settings", body)
	if err != nil {
		return err
	}
	return resp.Error()
}

// GetIndexMappings returns mappings for an index.
func (c *Client) GetIndexMappings(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/"+name+"/_mapping")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// PutIndexMappings updates mappings for an index.
func (c *Client) PutIndexMappings(ctx context.Context, name string, body interface{}) error {
	resp, err := c.Put(ctx, "/"+name+"/_mapping", body)
	if err != nil {
		return err
	}
	return resp.Error()
}

// GetIndexStats returns statistics for an index.
func (c *Client) GetIndexStats(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/"+name+"/_stats")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Rollover rolls over an alias to a new index.
func (c *Client) Rollover(ctx context.Context, alias string, body interface{}) (json.RawMessage, error) {
	resp, err := c.Post(ctx, "/"+alias+"/_rollover", body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Reindex copies documents from one index to another.
func (c *Client) Reindex(ctx context.Context, body interface{}) (json.RawMessage, error) {
	resp, err := c.Post(ctx, "/_reindex", body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// IndexExists checks if an index exists. Returns true if 200, false if 404.
func (c *Client) IndexExists(ctx context.Context, name string) (bool, error) {
	resp, err := c.Get(ctx, "/"+name)
	if err != nil {
		return false, err
	}
	if resp.HTTPResponse.StatusCode == 404 {
		resp.HTTPResponse.Body.Close()
		return false, nil
	}
	if resp.HTTPResponse.StatusCode >= 400 {
		return false, fmt.Errorf("unexpected status %d checking index %s", resp.HTTPResponse.StatusCode, name)
	}
	resp.HTTPResponse.Body.Close()
	return true, nil
}
