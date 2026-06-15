package es

import (
	"io"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type snapshotGetResponse struct {
	Snapshots []snapshotResponse `json:"snapshots"`
}

type snapshotResponse struct {
	Name              string   `json:"snapshot"`
	Indices           []string `json:"indices"`
	State             string   `json:"state"`
	StartTimeInMillis int64    `json:"start_time_in_millis"`
	EndTimeInMillis   int64    `json:"end_time_in_millis"`
}

func (c *Client) CreateSnapshotRepository(name string, settings RepositorySettings) error {
	body, err := newJSONReader(map[string]any{
		"type": "s3",
		"settings": map[string]any{
			"bucket":                 settings.Bucket,
			"region":                 settings.Region,
			"server_side_encryption": settings.ServerSideEncryption,
			"protocol":               settings.Protocol,
		},
	})
	if err != nil {
		return err
	}

	res, err := c.es.Snapshot.CreateRepository(name, body, func(r *esapi.SnapshotCreateRepositoryRequest) {
		r.Verify = boolPtr(true)
	})
	if err != nil {
		return err
	}

	return closeResponse(res)
}

func (c *Client) TakeSnapshot(repository, name string, indices []string) error {
	var body io.Reader
	if len(indices) > 0 {
		var err error
		body, err = newJSONReader(map[string]any{"indices": joinNames(indices)})
		if err != nil {
			return err
		}
	}

	res, err := c.es.Snapshot.Create(repository, name, func(r *esapi.SnapshotCreateRequest) {
		r.Body = body
		r.WaitForCompletion = boolPtr(false)
	})
	if err != nil {
		return err
	}

	return closeResponse(res)
}

func (c *Client) GetSnapshot(repository, name string) (*Snapshot, error) {
	res, err := c.es.Snapshot.Get(repository, []string{name})
	if err != nil {
		return nil, err
	}

	var payload snapshotGetResponse
	if err := decodeResponse(res, &payload); err != nil {
		return nil, err
	}
	if len(payload.Snapshots) == 0 {
		return nil, nil
	}

	converted := toSnapshot(payload.Snapshots[0])
	return &converted, nil
}

func (c *Client) GetSnapshots(repository string) ([]Snapshot, error) {
	res, err := c.es.Snapshot.Get(repository, []string{"_all"})
	if err != nil {
		return nil, err
	}

	var payload snapshotGetResponse
	if err := decodeResponse(res, &payload); err != nil {
		return nil, err
	}

	snapshots := make([]Snapshot, 0, len(payload.Snapshots))
	for _, snapshot := range payload.Snapshots {
		snapshots = append(snapshots, toSnapshot(snapshot))
	}
	return snapshots, nil
}

func (c *Client) RestoreSnapshot(repository, name string, opts RestoreOptions) error {
	body, err := newJSONReader(map[string]any{
		"ignore_unavailable":   opts.IgnoreUnavailable,
		"include_global_state": opts.IncludeGlobalState,
		"include_aliases":      opts.IncludeAliases,
		"rename_pattern":       opts.RenamePattern,
		"rename_replacement":   opts.RenameReplacement,
	})
	if err != nil {
		return err
	}

	res, err := c.es.Snapshot.Restore(repository, name, func(r *esapi.SnapshotRestoreRequest) {
		r.Body = body
		r.WaitForCompletion = boolPtr(false)
	})
	if err != nil {
		return err
	}

	return closeResponse(res)
}

func (c *Client) DeleteSnapshot(repository, name string) error {
	res, err := c.es.Snapshot.Delete(repository, []string{name})
	if err != nil {
		return err
	}

	return closeResponse(res)
}

func toSnapshot(snapshot snapshotResponse) Snapshot {
	return Snapshot{
		Name:      snapshot.Name,
		Indices:   snapshot.Indices,
		State:     snapshot.State,
		StartTime: timeFromMillis(snapshot.StartTimeInMillis),
		EndTime:   timeFromMillis(snapshot.EndTimeInMillis),
	}
}

func timeFromMillis(ms int64) time.Time {
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	joined := names[0]
	for _, name := range names[1:] {
		joined += "," + name
	}
	return joined
}
