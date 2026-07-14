package client

import (
	"context"
)

// CatAlias represents an alias from _cat/aliases.
type CatAlias struct {
	Alias         string `json:"alias"`
	Index         string `json:"index"`
	Filter        string `json:"filter"`
	RoutingIndex  string `json:"routing.index"`
	RoutingSearch string `json:"routing.search"`
	IsWriteIndex  string `json:"is_write_index"`
}

// ListAliases returns aliases from _cat/aliases, optionally filtered by the
// index they point at. The _cat/aliases path parameter filters by alias name,
// not index, so the index filter is applied client-side.
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

// CreateAlias creates an alias for an index.
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

// DeleteAlias removes an alias from an index.
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
