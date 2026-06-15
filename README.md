# esnap

Manage Elasticsearch snapshots and take a nap.

## Introduction

`esnap` is a tool to easily manage Elasticsearch snapshots and related tasks. It follows strong conventions, which is important to understand to get the best use of it:

* It uses environments, e.g. `development`, `staging`, `production`, etc.
* Environments must match the name of the repositories.
* Snapshots are named after the timestamp they were took.
* Indices are named as follows: `<environment>_<index_identifier>_<restored_timestamp><snapshot_name>`
  e.g. `development_users_2015062217505220160405080246`
* Aliases are named as follows: `<environment>_<index_identifier>`, e.g. `staging_events`
* Aliases always belong to indices that follow their pattern:
  * `staging_events` might be an alias of `staging_events_2015062317402220160405081241`
  * `development_users` an alias of `development_users_2015062217505220160405080246`
  * etc.
* When restoring an index, the suffix `<restored_timestamp><snapshot_name>` will be re-established.

`esnap` supports self-managed Elasticsearch 7.x and 8.x clusters using the snapshot repository APIs. Repository creation targets the Elasticsearch `repository-s3` plugin; AWS credentials are configured on the cluster side, not in `esnap`.

## Prerequisites

* Go 1.21+ installed and configured.
* A self-managed Elasticsearch 7.x or 8.x cluster.
* The Elasticsearch `repository-s3` plugin installed on the cluster if you want `esnap init` to create S3-backed repositories.
* Snapshot repository access configured on the Elasticsearch side.

## Installation

```sh
go install github.com/bebanjo/esnap@latest
```

## Tests

```sh
go test ./...
```

## Usage

```text
Usage:
  esnap [command]

Available Commands:
  cleanup     Cleanup unused indices
  init        Creates a new repository
  restore     Restore a snapshot
  rotate      Rotate snapshots
  take        Take a snapshot

Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action

Use "esnap [command] --help" for more information about a command.
```

### Init a repository

```text
It is required to specify destination, so a new repository
will be created under this name, with a bucket named like <BUCKET><destination>
where <BUCKET> is defined in the configuration.

Usage:
  esnap init [flags]

Global Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action
```

### Take a snapshot

```text
You are required to set a destination. It will create a snapshot
on the destination repository. If repository does not exist, you can create
it with the provided flag.

Usage:
  esnap take [flags]

Flags:
      --aliased             Take snapshot of indices with associated aliases only
  -a, --all                 Take snapshot of all indices. Otherwise, only those matching the destination
  -r, --create-repository   Create repository

Global Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action
```

### Restore a snapshot

```text
You are required to set an origin, destination, and snapshot name.
By default, it will fetch the given snapshot from the origin repository, creating
new indices out of the ones from the snapshot, and make a swap of the alias, removing
the old indices. If you use the fresh option, all indices and alias will be restored,
without a swap.

Usage:
  esnap restore [flags]

Flags:
  -f, --fresh             Do a full, fresh restore of all data
  -o, --origin string     Origin of the snapshot to restore
  -s, --snapshot string   Name of the snapshot to restore

Global Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action
```

### Rotate snapshots

```text
Removes snapshots older than the given age, where default is 30 days.
You are required to set a `destination` flag, which represents the
environment where your snapshots are stored.

Usage:
  esnap rotate [flags]

Flags:
  -a, --age int   Maximun age in days to keep snapshots (default 30)

Global Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action
```

### Cleanup indices

```text
It will find all indices that are not pointed by an alias.
Handle with care in case this is an expected scenario!

Usage:
  esnap cleanup [flags]

Global Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action
```

## Configuration

If you want to set a custom prefix for your repository and an availability zone where your snapshots will be stored, you need to set a configuration file at `$HOME/.esnap.yaml`.

```yaml
bucket: "this-bucket-"
AZ: "eu-west-1"
elasticsearch_url: "http://localhost:9200"
elasticsearch_username: "elastic"
elasticsearch_password: "changeme"
```

Defaults are `my-bucket` for `bucket`, `eu-west-1` for `AZ`, `http://localhost:9200` for `elasticsearch_url`, and empty credentials for `elasticsearch_username` / `elasticsearch_password`.

You can also configure Elasticsearch connectivity with environment variables:

* `ES_URL`
* `ES_USERNAME`
* `ES_PASSWORD`

## License

MIT
