package client

import (
	"context"
	"encoding/json"
)

// ClusterHealth represents the Elasticsearch cluster health response.
type ClusterHealth struct {
	ClusterName                 string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

// ClusterInfo represents the response from GET / (cluster root).
type ClusterInfo struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Version     struct {
		Number string `json:"number"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

// GetClusterHealth returns the cluster health status.
func (c *Client) GetClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	resp, err := c.Get(ctx, "/_cluster/health")
	if err != nil {
		return nil, err
	}
	var health ClusterHealth
	if err := resp.JSON(&health); err != nil {
		return nil, err
	}
	return &health, nil
}

// GetClusterStats returns the cluster stats.
func (c *Client) GetClusterStats(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/_cluster/stats")
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetClusterSettings returns the cluster settings.
func (c *Client) GetClusterSettings(ctx context.Context, includeDefaults bool) (json.RawMessage, error) {
	path := "/_cluster/settings"
	if includeDefaults {
		path += "?include_defaults=true"
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

// PendingTask represents a pending cluster task.
type PendingTask struct {
	InsertOrder    int    `json:"insert_order"`
	Priority       string `json:"priority"`
	Source         string `json:"source"`
	TimeInQueueMS  int    `json:"time_in_queue_millis"`
	TimeInQueue    string `json:"time_in_queue"`
}

// GetPendingTasks returns the list of pending cluster tasks.
func (c *Client) GetPendingTasks(ctx context.Context) ([]PendingTask, error) {
	resp, err := c.Get(ctx, "/_cluster/pending_tasks")
	if err != nil {
		return nil, err
	}
	var result struct {
		Tasks []PendingTask `json:"tasks"`
	}
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result.Tasks, nil
}

// AllocationExplain returns the allocation explanation for a shard.
func (c *Client) AllocationExplain(ctx context.Context, body interface{}) (json.RawMessage, error) {
	resp, err := c.Post(ctx, "/_cluster/allocation/explain", body)
	if err != nil {
		return nil, err
	}
	var result json.RawMessage
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetClusterInfo returns basic cluster information from GET /.
func (c *Client) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	resp, err := c.Get(ctx, "/")
	if err != nil {
		return nil, err
	}
	var info ClusterInfo
	if err := resp.JSON(&info); err != nil {
		return nil, err
	}
	return &info, nil
}
