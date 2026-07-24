package es

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Client struct {
	es *elasticsearch.Client
}

type Config struct {
	Addresses []string
	Username  string
	Password  string
}

func NewClient(cfg Config) (*Client, error) {
	addresses := make([]string, 0, len(cfg.Addresses))
	for _, address := range cfg.Addresses {
		address = strings.TrimSpace(address)
		if address != "" {
			addresses = append(addresses, address)
		}
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("at least one elasticsearch address is required")
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:               addresses,
		Username:                cfg.Username,
		Password:                cfg.Password,
		EnableCompatibilityMode: true,
	})
	if err != nil {
		return nil, err
	}

	return &Client{es: esClient}, nil
}

func boolPtr(v bool) *bool {
	return &v
}

func newJSONReader(payload any) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}
	return &buf, nil
}

func closeResponse(res *esapi.Response) error {
	if res == nil {
		return nil
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		trimmed := strings.TrimSpace(string(body))
		if trimmed == "" {
			return errors.New(res.Status())
		}
		return fmt.Errorf("%s: %s", res.Status(), trimmed)
	}
	_, _ = io.Copy(io.Discard, res.Body)
	return nil
}

func decodeResponse(res *esapi.Response, out any) error {
	if res == nil {
		return fmt.Errorf("empty response")
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		trimmed := strings.TrimSpace(string(body))
		if trimmed == "" {
			return errors.New(res.Status())
		}
		return fmt.Errorf("%s: %s", res.Status(), trimmed)
	}
	return json.NewDecoder(res.Body).Decode(out)
}
