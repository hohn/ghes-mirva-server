package common

import "github.com/hohn/ghes-mirva-server/api"

type AnalyzeResult struct {
	RunAnalysisOutput string
}

type OwnerRepo struct {
	Owner string
	Repo  string
}

type B64GzTar struct {
	TGZFilepath string
}

type DBLocation struct {
	DBPATH string
}

type DBLocationLocal struct {
	prefix  string
	db_file string
}

type JobInfo struct {
	QueryLanguage string
	CreatedAt     string
	UpdatedAt     string

	Status              Status
	SkippedRepositories api.SkippedRepositories
}
type JobSpec struct {
	ID  int
	Orl OwnerRepo
}

type Status int

const (
	StatusInProgress = iota
	StatusQueued
	StatusError
	StatusSuccess
	StatusFailed
)

func (s Status) StatusString() string {
	switch s {
	case StatusInProgress:
		return "InProgress"
	case StatusQueued:
		return "Queued"
	case StatusError:
		return "Error"
	case StatusSuccess:
		return "Success"
	case StatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}
