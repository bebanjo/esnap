# esnap

Manage Elasticsearch snapshots and take a nap.

## Introduction

`esnap` is a tool to easily manage Elasticsearch snapshots and another
related tasks. It follows strong conventions, which is important to understand
to get the best use of it:

* It uses environments, e.g. `development`, `staging`, `production`, etc.
* Snapshots are named after the timestamp they were took.
* Indices are named as follows: `<environment>_<index_identifier>_<restored_timestamp><snapshot_name>`
  e.g. `development_users_2015062217505220160405080246`
* Aliases are named as follows: `<environment>_<index_identifier>`, e.g. `staging_events`
* Aliases always belong to indices that follow their pattern:
  * `staging_events` might be an alias of `staging_events_2015062317402220160405081241`
  * `development_users` an alias of `development_users_2015062217505220160405080246`
  * etc.
* When restoring an index, the suffix `<restored_timestamp><snapshot_name>` will be re-established.

For now, it is only compatible with Elasticsearch 1.X version and S3 snapshots.

## Prerequisites

* Go 1.6+ installed and configured.
* Configuration set to S3 on your end.
* Elasticsearch 1.X
* `elasticsearch-cloud-aws` plugin installed.

## Installation

`go get github.com/bebanjo/esnap`

## Tests

`go test ./...`

## Usage

```
Usage:
  esnap [command]

Available Commands:
  cleanup     Cleanup unused indices
  init        Creates a new repository
  restore     Restore a snapshot
  take        Take a snapshot

Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action

Use "esnap [command] --help" for more information about a command.
```

### Init a repository

```
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

```
You are required to set a destination. It will create a snapshot
on the destination repository. If repository does not exist, you can create
it with the provided flag.

Usage:
  esnap take [flags]

Flags:
  -a, --all                 Take snapshot of all indices. Otherwise, only those matching the destination
  -r, --create-repository   Create repository

Global Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action

```

### Restore a snapshot

```
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

### Cleanup indices

```
It will find all indices that are not pointed by an alias.
Handle with care in case this is an expected scenario!

Usage:
  esnap cleanup [flags]

Global Flags:
      --config string        config file (default is $HOME/.esnap.yaml)
  -d, --destination string   Destination for the command action
```

## License

MIT
