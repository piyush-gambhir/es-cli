package client

import (
	"context"
)

// CatShard represents a shard from _cat/shards.
type CatShard struct {
	Index  string `json:"index"`
	Shard  string `json:"shard"`
	PriRep string `json:"prirep"`
	State  string `json:"state"`
	Docs   string `json:"docs"`
	Store  string `json:"store"`
	IP     string `json:"ip"`
	Node   string `json:"node"`
}

// ListShards returns shards from _cat/shards.
func (c *Client) ListShards(ctx context.Context, index string) ([]CatShard, error) {
	path := "/_cat/shards?format=json&h=index,shard,prirep,state,docs,store,ip,node"
	if index != "" {
		path = "/_cat/shards/" + index + "?format=json&h=index,shard,prirep,state,docs,store,ip,node"
	}
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var shards []CatShard
	if err := resp.JSON(&shards); err != nil {
		return nil, err
	}
	return shards, nil
}
