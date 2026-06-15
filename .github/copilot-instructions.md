# esnap – Copilot Instructions

## Build & test

```bash
go build ./...                      # build
go test ./...                       # full test suite
go test ./internal/es/...           # test ES client package only
go test ./cmd/...                   # test command helpers only
go test -run TestCreateSnapshotRepository ./internal/es/  # single test
```

## Architecture

`esnap` is a CLI tool (cobra/viper) that manages Elasticsearch snapshots via the ES REST API.

```
main.go              → cmd.Execute()
cmd/                 → cobra commands (root, init, take, restore, rotate, cleanup)
internal/es/         → ES client abstraction
  client.go          → NewClient, shared HTTP helpers (decodeResponse, closeResponse)
  snapshot.go        → snapshot repository + snapshot CRUD
  index.go           → _cat/indices, _cat/aliases, alias updates, index deletion
  types.go           → shared domain types (Snapshot, Index, Alias, RepositorySettings, RestoreOptions)
```

**Key flow**: Each cobra command calls `mustClient()` (from `cmd/root.go`) to get the singleton `internal/es.Client`, then calls methods on it directly. There is no service layer between commands and the client.

**ES compatibility**: `go-elasticsearch/v8` with `EnableCompatibilityMode: true` — this makes it work against both ES 7.x and 8.x. The response header `X-Elastic-Product: Elasticsearch` must be present (mocked in tests via `newTestClient`).

## Key conventions

**Naming**: Indices follow `<env>_<identifier>_<timestamp><snapshot>` (e.g. `staging_users_20240101120000`). Aliases follow `<env>_<identifier>` (e.g. `staging_users`). The `--destination` flag is the environment name and also the snapshot repository name.

**Cat API format**: Always request `format=json` (not text) when using `_cat/indices` and `_cat/aliases`. The old library parsed text; this codebase decodes JSON.

**Snapshot timestamps**: Named as `time.Now().Format("20060102150405")`. `StartTime`/`EndTime` come from `start_time_in_millis` / `end_time_in_millis` fields, not the time strings.

**Error handling pattern**: Commands print to `os.Stderr` and call `os.Exit(1)`. The `internal/es` package only returns errors; it never exits.

**Config resolution order** (viper): env vars override config file, which overrides defaults.
- `ES_URL` → `elasticsearch_url` (default: `http://localhost:9200`, comma-separated for multiple nodes)
- `ES_USERNAME` → `elasticsearch_username`
- `ES_PASSWORD` → `elasticsearch_password`
- `bucket`, `AZ`, `protocol`, `server_side_encryption` in `$HOME/.esnap.yaml` (defaults: `my-bucket`, `eu-west-1`, `https`, `true`)

**Testing**: `internal/es` tests use `httptest.NewServer` with a handler passed to `newTestClient`. The test server always adds `X-Elastic-Product: Elasticsearch` (required by the ES client). Basic auth is asserted with `assertBasicAuth`. `cmd/` tests only test pure helper functions (no mocked ES calls needed there).

**ES 7.x / 8.x compatibility**: The go-elasticsearch/v8 client adds some query parameters that ES 7.x rejects (e.g. `index_names`, `wait_for_completion` on DELETE). Strip any such options from client calls; the ES 7.x defaults are correct. Use `EnableCompatibilityMode: true` in `NewClient`.
