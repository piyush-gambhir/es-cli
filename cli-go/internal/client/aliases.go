package client

import (
	"context"
)

type CatAlias struct {
	Alias         string `json:"alias"`
	Index         string `json:"index"`
	Filter        string `json:"filter"`
	RoutingIndex  string `json:"routing.index"`
	RoutingSearch string `json:"routing.search"`
	IsWriteIndex  string `json:"is_write_index"`
}

func (c *Client) ListAliases(ctx context.Context, index string) ([]CatAlias, error) {
	resp, err := c.Get(ctx, "/_cat/aliases?format=json")
	if err != nil {
		return nil, err
	}
	var aliases []CatAlias
	if err := resp.JSON(&aliases); err != nil {
		return nil, err
	}
	if index != "" {
		filtered := make([]CatAlias, 0, len(aliases))
		for _, a := range aliases {
			if a.Index == index {
				filtered = append(filtered, a)
			}
		}
		aliases = filtered
	}
	return aliases, nil
}

func (c *Client) CreateAlias(ctx context.Context, index, alias string) error {
	body := map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"add": map[string]interface{}{
					"index": index,
					"alias": alias,
				},
			},
		},
	}
	resp, err := c.Post(ctx, "/_aliases", body)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DeleteAlias(ctx context.Context, index, alias string) error {
	body := map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"remove": map[string]interface{}{
					"index": index,
					"alias": alias,
				},
			},
		},
	}
	resp, err := c.Post(ctx, "/_aliases", body)
	if err != nil {
		return err
	}
	return resp.Error()
}
