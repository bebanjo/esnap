package es

import "github.com/elastic/go-elasticsearch/v8/esapi"

type catIndexResponse struct {
	Health string `json:"health"`
	Status string `json:"status"`
	Index  string `json:"index"`
}

type catAliasResponse struct {
	Alias string `json:"alias"`
	Index string `json:"index"`
}

func (c *Client) GetIndices(pattern string) ([]Index, error) {
	res, err := c.es.Cat.Indices(func(r *esapi.CatIndicesRequest) {
		if pattern != "" {
			r.Index = []string{pattern}
		}
		r.Format = "json"
		r.H = []string{"health", "status", "index"}
	})
	if err != nil {
		return nil, err
	}

	var payload []catIndexResponse
	if err := decodeResponse(res, &payload); err != nil {
		return nil, err
	}

	indices := make([]Index, 0, len(payload))
	for _, item := range payload {
		indices = append(indices, Index{Name: item.Index, Health: item.Health, Status: item.Status})
	}
	return indices, nil
}

func (c *Client) GetAliases(pattern string) ([]Alias, error) {
	res, err := c.es.Cat.Aliases(func(r *esapi.CatAliasesRequest) {
		if pattern != "" {
			r.Name = []string{pattern}
		}
		r.Format = "json"
		r.H = []string{"alias", "index"}
	})
	if err != nil {
		return nil, err
	}

	var payload []catAliasResponse
	if err := decodeResponse(res, &payload); err != nil {
		return nil, err
	}

	aliases := make([]Alias, 0, len(payload))
	for _, item := range payload {
		aliases = append(aliases, Alias{Name: item.Alias, Index: item.Index})
	}
	return aliases, nil
}

func (c *Client) AddAlias(index, alias string) error {
	body, err := newJSONReader(map[string]any{
		"actions": []map[string]any{{
			"add": map[string]string{
				"index": index,
				"alias": alias,
			},
		}},
	})
	if err != nil {
		return err
	}

	res, err := c.es.Indices.UpdateAliases(body)
	if err != nil {
		return err
	}

	return closeResponse(res)
}

func (c *Client) DeleteIndex(name string) error {
	res, err := c.es.Indices.Delete([]string{name})
	if err != nil {
		return err
	}

	return closeResponse(res)
}
