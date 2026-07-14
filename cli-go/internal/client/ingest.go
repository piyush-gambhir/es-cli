package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListPipelines returns all ingest pipelines.
func (c *Client) ListPipelines(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/_ingest/pipeline")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPipeline returns a specific ingest pipeline by name.
func (c *Client) GetPipeline(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/_ingest/pipeline/%s", name))
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreatePipeline creates or updates an ingest pipeline.
func (c *Client) CreatePipeline(ctx context.Context, name string, body interface{}) error {
	resp, err := c.Put(ctx, fmt.Sprintf("/_ingest/pipeline/%s", name), body)
	if err != nil {
		return err
	}
	return resp.Error()
}

// DeletePipeline deletes an ingest pipeline by name.
func (c *Client) DeletePipeline(ctx context.Context, name string) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/_ingest/pipeline/%s", name))
	if err != nil {
		return err
	}
	return resp.Error()
}

// SimulatePipeline simulates an ingest pipeline with the given documents.
func (c *Client) SimulatePipeline(ctx context.Context, name string, body interface{}) (json.RawMessage, error) {
	resp, err := c.Post(ctx, fmt.Sprintf("/_ingest/pipeline/%s/_simulate", name), body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}
