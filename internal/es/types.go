package es

import "time"

type RepositorySettings struct {
	Bucket               string
	Region               string
	ServerSideEncryption bool
	Protocol             string
}

type Snapshot struct {
	Name      string
	Indices   []string
	State     string
	StartTime time.Time
	EndTime   time.Time
}

type Index struct {
	Name   string
	Health string
	Status string
}

type Alias struct {
	Name  string
	Index string
}

type RestoreOptions struct {
	IgnoreUnavailable  bool
	IncludeGlobalState bool
	IncludeAliases     bool
	RenamePattern      string
	RenameReplacement  string
}
