package common

type AnalyzeResult struct {
	RunAnalysisOutput string
}

type OwnerRepoLoc struct {
	Owner    string
	Repo     string
	Location DBLocation
}

type B64GzTar struct {
	TGZFilepath string
}

type DBLocation interface {
	DBPATH() string
}

type DBLocationLocal struct {
	prefix  string
	db_file string
}
