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
