package client

import (
	"context"
	"encoding/json"
)

// CatNode represents a node from _cat/nodes.
type CatNode struct {
	IP          string `json:"ip"`
	HeapPercent string `json:"heap.percent"`
	RAMPercent  string `json:"ram.percent"`
	CPU         string `json:"cpu"`
	Load1m      string `json:"load_1m"`
	Load5m      string `json:"load_5m"`
	Load15m     string `json:"load_15m"`
	NodeRole    string `json:"node.role"`
	Master      string `json:"master"`
	Name        string `json:"name"`
}

// ListNodes returns nodes from _cat/nodes.
func (c *Client) ListNodes(ctx context.Context) ([]CatNode, error) {
	resp, err := c.Get(ctx, "/_cat/nodes?format=json&h=ip,heap.percent,ram.percent,cpu,load_1m,load_5m,load_15m,node.role,master,name")
	if err != nil {
		return nil, err
	}
	var nodes []CatNode
	if err := resp.JSON(&nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// GetNodeInfo returns detailed information about a specific node.
func (c *Client) GetNodeInfo(ctx context.Context, nodeID string) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/_nodes/"+nodeID)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetNodeStats returns statistics for nodes.
func (c *Client) GetNodeStats(ctx context.Context, nodeID, metric string) (json.RawMessage, error) {
	path := "/_nodes"
	if nodeID != "" {
		path += "/" + nodeID
	}
	path += "/stats"
	if metric != "" {
		path += "/" + metric
	}
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetHotThreads returns hot threads information. Returns plain text.
func (c *Client) GetHotThreads(ctx context.Context, nodeID string) (string, error) {
	path := "/_nodes"
	if nodeID != "" {
		path += "/" + nodeID
	}
	path += "/hot_threads"
	resp, err := c.Get(ctx, path)
	if err != nil {
		return "", err
	}
	body, err := resp.RawBody()
	if err != nil {
		return "", err
	}
	return string(body), nil
}
