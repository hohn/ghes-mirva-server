package common

import "github.com/hohn/ghes-mirva-server/api"

type AnalyzeResult struct {
	RunAnalysisSARIF string
	RunAnalysisBQRS  string
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

	SkippedRepositories api.SkippedRepositories
}

type JobSpec struct {
	ID int
	OwnerRepo
}

type Status int

const (
	StatusInProgress = iota
	StatusQueued
	StatusError
	StatusSuccess
	StatusFailed
)

type AnalyzeJob struct {
	MirvaRequestID int

	QueryPackId   int
	QueryLanguage string

	ORL OwnerRepo
}

func StatusFromString(status string) Status {
	switch status {
	case "in_progress":
		return StatusInProgress
	case "InProgress":
		return StatusInProgress
	case "Queued":
		return StatusQueued
	case "Error":
		return StatusError
	case "Success":
		return StatusSuccess
	case "Failed":
		return StatusFailed
	default:
		return -1 // or handle the invalid status case accordingly
	}
}

func (s Status) ToExternalString() string {
	switch s {
	case StatusInProgress:
		return "in_progress"
	case StatusQueued:
		return "queued"
	case StatusError:
		return "error"
	case StatusSuccess:
		return "succeeded"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}
