package analyze

import (
	"log/slog"
)

var (
	NumWorkers int
	Jobs       chan AnalyzeJob
	Results    chan AnalyzeResult
)

type AnalyzeJob struct {
	id int
}

type AnalyzeResult struct {
	id int
}

func worker(wid int, jobs <-chan AnalyzeJob, results chan<- AnalyzeResult) {
	var job AnalyzeJob
	for {
		job = <-jobs
		slog.Debug("Picking up job", "job", job, "worker", wid)

		// And run it

		// TODO For every repository with a built database we ultimately run
		// the queries to create the sarif file.  See
		//        cmd/run-analyze.sh
		// for details

		// results <- job * 2
	}
}

func RunWorkers() {
	// TODO as cli arg

	// for a := 1; a <= numJobs; a++ {
	// 	<-results
	// }
}

func init() {
	Jobs = make(chan AnalyzeJob)
	Results = make(chan AnalyzeResult)
	NumWorkers = 2

	for wid := 1; wid <= NumWorkers; wid++ {
		go worker(wid, Jobs, Results)
	}
}
