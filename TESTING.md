# Testing esnap

This document describes how to run the unit tests, smoke-test against a real cluster via SSH tunnel, and run a full end-to-end integration test locally using Docker.

---

## Unit tests

```sh
go test ./...
```

---

## Smoke tests against a real cluster (read-only)

These use an SSH tunnel to reach a private ES node. No data is written or deleted.

### 1. Open an SSH tunnel

```sh
# Replace staging-elastic01-<id>.bebanjo.net with the actual hostname from AWS EC2
ssh -fN -L 19200:<private-ip>:9200 bastion
```

Or use the SSH config shorthand if your `~/.ssh/config` has a `ProxyCommand` for `*.bebanjo.net`:

```sh
ssh -fN -L 19200:staging-elastic01-<id>.bebanjo.net:9200 bastion
```

### 2. Verify connectivity

```sh
curl http://localhost:19200/
# Expect: {"name": ..., "version": {"number": "7.17.x"}, ...}
```

### 3. Build and run read-only commands

```sh
go build -o esnap .

# Verify rotate logic without deleting anything (age=9999 days)
ES_URL=http://localhost:19200 ./esnap rotate --destination staging --age 9999
# Expected output: "Found N snapshots on staging", "0 snapshots on staging were rotated"
```

---

## Full local integration test (Docker)

Exercises every command path including writes, snapshot, restore, and cleanup.

### Prerequisites

- Docker
- AWS CLI (`aws`)
- `go` 1.21+

### 1. Build the ES Docker image

The official ES 7.17 image does not include the `repository-s3` plugin and has a JVM cgroup v2 compatibility bug on Linux kernels >= 5.x. The steps below address both.

```sh
mkdir -p /tmp/esnap-test

# Download the S3 plugin locally (avoids SSL issues in the Docker build network)
curl -fL -o /tmp/esnap-test/repository-s3.zip \
  "https://artifacts.elastic.co/downloads/elasticsearch-plugins/repository-s3/repository-s3-7.17.2.zip"

cat > /tmp/esnap-test/Dockerfile.es << 'EOF'
FROM docker.elastic.co/elasticsearch/elasticsearch:7.17.2
COPY repository-s3.zip /tmp/repository-s3.zip
RUN elasticsearch-plugin install --batch file:///tmp/repository-s3.zip && rm /tmp/repository-s3.zip
# Fix cgroup v2 JVM bug: inject -XX:-UseContainerSupport into the JvmOptionsParser
# invocation so the launcher doesn't crash on hosts using the unified cgroup hierarchy.
RUN sed -i \
  's|ES_JAVA_OPTS=`export ES_TMPDIR; "$JAVA" "$XSHARE"|ES_JAVA_OPTS=`export ES_TMPDIR; "$JAVA" -XX:-UseContainerSupport "$XSHARE"|' \
  /usr/share/elasticsearch/bin/elasticsearch
EOF

docker build -f /tmp/esnap-test/Dockerfile.es -t esnap-es:test /tmp/esnap-test
```

### 2. Write the ES config

```sh
cat > /tmp/esnap-test/elasticsearch.yml << 'EOF'
cluster.name: esnap-test
discovery.type: single-node
network.host: 0.0.0.0
xpack.security.enabled: false
s3.client.default.endpoint: esnap-minio:9000
s3.client.default.protocol: http
s3.client.default.path_style_access: true
EOF
```

### 3. Start MinIO and ES

```sh
docker network create esnap-test 2>/dev/null || true

# MinIO — local S3-compatible storage
docker run -d --name esnap-minio \
  --network esnap-test \
  -p 9100:9000 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data

# ES 7.17.2 with S3 plugin
# Note: ES_JAVA_OPTS also carries -XX:-UseContainerSupport for the actual ES JVM.
docker run -d --name esnap-es \
  --network esnap-test \
  -p 19300:9200 \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m -XX:-UseContainerSupport" \
  -v /tmp/esnap-test/elasticsearch.yml:/usr/share/elasticsearch/config/elasticsearch.yml \
  esnap-es:test

# Wait for ES to be healthy (takes ~30s)
until curl -s http://localhost:19300/_cluster/health | grep -q '"status":"green"\|"status":"yellow"'; do
  echo "waiting for ES..."; sleep 5
done
echo "ES ready"
```

### 4. Configure credentials and create the MinIO bucket

The ES S3 plugin uses its own credential chain and does **not** fall through to `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` env vars. Credentials must be in the ES keystore.

```sh
# Add MinIO credentials to the ES keystore
echo "minioadmin" | docker exec -i esnap-es bin/elasticsearch-keystore add --stdin s3.client.default.access_key
echo "minioadmin" | docker exec -i esnap-es bin/elasticsearch-keystore add --stdin s3.client.default.secret_key

# Reload secure settings without restarting ES
curl -s -X POST http://localhost:19300/_nodes/reload_secure_settings | python3 -m json.tool

# Create the MinIO bucket (AWS CLI requires SigV4 so use --endpoint-url)
AWS_ACCESS_KEY_ID=minioadmin AWS_SECRET_ACCESS_KEY=minioadmin \
  aws s3 mb s3://test-testenv \
  --endpoint-url http://localhost:9100 \
  --region us-east-1
```

### 5. Write the esnap config

```sh
cat > /tmp/esnap-test/.esnap.yaml << 'EOF'
elasticsearch_url: "http://localhost:19300"
bucket: "test-"
AZ: "us-east-1"
protocol: "http"
server_side_encryption: false
EOF
```

`bucket: "test-"` + `destination: "testenv"` → MinIO bucket `test-testenv`.

### 6. Build esnap

```sh
cd /path/to/esnap
go build -o esnap .
ESNAP="./esnap --config /tmp/esnap-test/.esnap.yaml"
ES=http://localhost:19300
```

### 7. Seed test data

```sh
# Three aliased indices (normal data)
for name in users products orders; do
  IDX="testenv_${name}_20240101120000"
  curl -s -X PUT "$ES/$IDX" -H 'Content-Type: application/json' \
    -d '{"settings":{"number_of_replicas":0}}' > /dev/null
  curl -s -X POST "$ES/_aliases" -H 'Content-Type: application/json' \
    -d "{\"actions\":[{\"add\":{\"index\":\"$IDX\",\"alias\":\"testenv_$name\"}}]}" > /dev/null
  echo "seeded $IDX + alias testenv_$name"
done

# Two orphan indices (no alias — for the cleanup test)
for name in orphan_a orphan_b; do
  IDX="testenv_${name}_20240101120000"
  curl -s -X PUT "$ES/$IDX" -H 'Content-Type: application/json' \
    -d '{"settings":{"number_of_replicas":0}}' > /dev/null
  echo "seeded $IDX (orphan, no alias)"
done
```

### 8. Run all command paths

#### `init` — create the S3 snapshot repository

```sh
$ESNAP init -d testenv
# Expected: "creating repository testenv" with no error
```

#### `take` — take a snapshot and wait for completion

```sh
$ESNAP take -d testenv
# Expected: polls until "SUCCESS"

SNAPSHOT=$(curl -s "$ES/_snapshot/testenv/_all" | \
  python3 -c "import json,sys; print(json.load(sys.stdin)['snapshots'][-1]['snapshot'])")
echo "Snapshot name: $SNAPSHOT"
```

#### `rotate` — delete old snapshots

```sh
# Safe check: age=9999 deletes nothing
$ESNAP rotate -d testenv --age 9999
# Expected: "Found N snapshots on testenv", "0 snapshots on testenv were rotated"

# Active delete: age=0 deletes everything (snapshots older than now)
$ESNAP rotate -d testenv --age 0
# Expected: "Removed snapshot <name> from testenv" for each snapshot
```

> **Note:** Rotate uses the snapshot's actual `start_time_in_millis`, not its name.
> Snapshots just created will not be deleted by `--age 1` (they are less than 1 day old).

#### `restore --fresh` — restore into a new environment (no alias swap)

```sh
# Take a snapshot first if you ran rotate above
$ESNAP take -d testenv
SNAPSHOT=$(curl -s "$ES/_snapshot/testenv/_all" | \
  python3 -c "import json,sys; print(json.load(sys.stdin)['snapshots'][-1]['snapshot'])")

$ESNAP restore -d testenv3 -o testenv -s $SNAPSHOT --fresh
# Expected: indices like testenv3_users_<date><snapshot> appear, no aliases created
curl -s "$ES/_cat/indices/testenv3_*?v&h=index,health"
```

#### `restore` (swap) — replace an existing environment's aliases

```sh
# Pre-seed "old" testenv4 indices with aliases
for name in users products; do
  IDX="testenv4_${name}_20230101000000"
  curl -s -X PUT "$ES/$IDX" -H 'Content-Type: application/json' \
    -d '{"settings":{"number_of_replicas":0}}' > /dev/null
  curl -s -X POST "$ES/_aliases" -H 'Content-Type: application/json' \
    -d "{\"actions\":[{\"add\":{\"index\":\"$IDX\",\"alias\":\"testenv4_$name\"}}]}" > /dev/null
done

$ESNAP restore -d testenv4 -o testenv -s $SNAPSHOT
# Expected: new testenv4_* indices created, aliases swapped, old indices deleted
curl -s "$ES/_cat/aliases/testenv4_*?v&h=alias,index"
```

#### `cleanup` — delete all unaliased indices

```sh
$ESNAP cleanup
# Expected: orphan indices deleted, aliased indices survive
curl -s "$ES/_cat/indices?v&h=index,health" | sort
curl -s "$ES/_cat/aliases?v&h=alias,index" | sort
```

### 9. Tear down

```sh
docker rm -f esnap-es esnap-minio
docker network rm esnap-test
rm -rf /tmp/esnap-test
```

---

## Known ES 7.x compatibility notes

| Issue | Root cause | Fix applied |
|-------|-----------|-------------|
| `GET _snapshot` returns 400 | `?index_names=true` is ES 8.x-only | Removed from `GetSnapshot`/`GetSnapshots` |
| `DELETE _snapshot` returns 400 | `?wait_for_completion=true` not accepted in ES 7.x | Removed from `DeleteSnapshot` |
| S3 repository creation with `protocol: https` fails against HTTP endpoints | Protocol was hardcoded to `https` | Now configurable via `protocol:` config key (default `https`) |
| ES S3 plugin ignores `AWS_ACCESS_KEY_ID` env var | Plugin uses its own credential chain, not the AWS SDK default chain | Must use ES keystore: `elasticsearch-keystore add s3.client.default.access_key` |
| ES 7.17 Docker fails on cgroup v2 hosts | JVM cgroup v2 detection bug in OpenJDK bundled with 7.17 | Patched `elasticsearch` startup script to pass `-XX:-UseContainerSupport` to `JvmOptionsParser`; also pass it in `ES_JAVA_OPTS` |
