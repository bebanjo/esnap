package es

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateSnapshotRepository(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertBasicAuth(t, r)
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/_snapshot/staging" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("verify"); got != "true" {
			t.Fatalf("unexpected verify query %q", got)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		settings := payload["settings"].(map[string]any)
		if payload["type"] != "s3" || settings["bucket"] != "my-bucket-staging" || settings["region"] != "eu-west-1" {
			t.Fatalf("unexpected payload %#v", payload)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"acknowledged":true}`))
	})
	defer server.Close()

	if err := client.CreateSnapshotRepository("staging", RepositorySettings{
		Bucket:               "my-bucket-staging",
		Region:               "eu-west-1",
		ServerSideEncryption: true,
		Protocol:             "https",
	}); err != nil {
		t.Fatalf("CreateSnapshotRepository() error = %v", err)
	}
}

func TestSnapshotAndCatOperations(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertBasicAuth(t, r)
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/_snapshot/staging/20240101120000":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"indices":"index_a,index_b"`) {
				t.Fatalf("unexpected snapshot body %s", body)
			}
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"accepted":true}`))
		case r.Method == http.MethodGet && r.URL.Path == "/_snapshot/staging/20240101120000":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"snapshots":[{"snapshot":"20240101120000","indices":["index_a","index_b"],"state":"SUCCESS","start_time_in_millis":1710000000000,"end_time_in_millis":1710000300000}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/_snapshot/staging/_all":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"snapshots":[{"snapshot":"old","indices":["index_a"],"state":"SUCCESS","start_time_in_millis":1700000000000,"end_time_in_millis":1700000300000}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/_cat/indices/staging*":
			if got := r.URL.Query().Get("format"); got != "json" {
				t.Fatalf("unexpected format %q", got)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"health":"green","status":"open","index":"staging_users_1"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/_cat/aliases/staging*":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"alias":"staging_users","index":"staging_users_1"}]`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	if err := client.TakeSnapshot("staging", "20240101120000", []string{"index_a", "index_b"}); err != nil {
		t.Fatalf("TakeSnapshot() error = %v", err)
	}

	snapshot, err := client.GetSnapshot("staging", "20240101120000")
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}
	if snapshot == nil || snapshot.Name != "20240101120000" || snapshot.State != "SUCCESS" {
		t.Fatalf("unexpected snapshot %#v", snapshot)
	}

	snapshots, err := client.GetSnapshots("staging")
	if err != nil {
		t.Fatalf("GetSnapshots() error = %v", err)
	}
	if len(snapshots) != 1 || snapshots[0].Name != "old" {
		t.Fatalf("unexpected snapshots %#v", snapshots)
	}

	indices, err := client.GetIndices("staging*")
	if err != nil {
		t.Fatalf("GetIndices() error = %v", err)
	}
	if len(indices) != 1 || indices[0].Name != "staging_users_1" {
		t.Fatalf("unexpected indices %#v", indices)
	}

	aliases, err := client.GetAliases("staging*")
	if err != nil {
		t.Fatalf("GetAliases() error = %v", err)
	}
	if len(aliases) != 1 || aliases[0].Name != "staging_users" {
		t.Fatalf("unexpected aliases %#v", aliases)
	}
	if snapshot.StartTime != time.UnixMilli(1710000000000).UTC() {
		t.Fatalf("unexpected start time %v", snapshot.StartTime)
	}
}

func TestRestoreAliasAndDeleteOperations(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertBasicAuth(t, r)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/_snapshot/staging/snap-1/_restore":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"include_aliases":false`) || !strings.Contains(string(body), `"rename_pattern":"staging_(.+)_\\d+(_.*)?"`) {
				t.Fatalf("unexpected restore body %s", body)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"accepted":true}`))
		case r.Method == http.MethodPost && r.URL.Path == "/_aliases":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"alias":"staging_users"`) || !strings.Contains(string(body), `"index":"staging_users_2"`) {
				t.Fatalf("unexpected alias body %s", body)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/_snapshot/staging/snap-1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/staging_users_1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	})
	defer server.Close()

	if err := client.RestoreSnapshot("staging", "snap-1", RestoreOptions{
		IgnoreUnavailable:  true,
		IncludeGlobalState: false,
		IncludeAliases:     false,
		RenamePattern:      `staging_(.+)_\d+(_.*)?`,
		RenameReplacement:  `restored_$1_suffix`,
	}); err != nil {
		t.Fatalf("RestoreSnapshot() error = %v", err)
	}
	if err := client.AddAlias("staging_users_2", "staging_users"); err != nil {
		t.Fatalf("AddAlias() error = %v", err)
	}
	if err := client.DeleteSnapshot("staging", "snap-1"); err != nil {
		t.Fatalf("DeleteSnapshot() error = %v", err)
	}
	if err := client.DeleteIndex("staging_users_1"); err != nil {
		t.Fatalf("DeleteIndex() error = %v", err)
	}
}

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		handler(w, r)
	}))
	client, err := NewClient(Config{
		Addresses: []string{server.URL},
		Username:  "user",
		Password:  "pass",
	})
	if err != nil {
		server.Close()
		t.Fatalf("NewClient() error = %v", err)
	}
	return client, server
}

func assertBasicAuth(t *testing.T, r *http.Request) {
	t.Helper()
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if got := r.Header.Get("Authorization"); got != want {
		t.Fatalf("unexpected authorization %q", got)
	}
}
