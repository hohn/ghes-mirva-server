package analyze

import (
	"bufio"
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

var (
	NumWorkers int
	Jobs       chan AnalyzeJob
	Results    chan AnalyzeResult
)

type AnalyzeJob struct {
	QueryPackId   int
	QueryLanguage string

	DBOwner string
	DBRepo  string
}

type AnalyzeResult struct {
	RunAnalysisOutput string
}

func worker(wid int, jobs <-chan AnalyzeJob, results chan<- AnalyzeResult) {
	var job AnalyzeJob
	for {
		job = <-jobs
		slog.Debug("Picking up job", "job", job, "worker", wid)

		cwd, err := os.Getwd()
		if err != nil {
			slog.Error("RunJob: cwd problem: ", "error", err)
			continue
		}

		slog.Debug("Analysis: running", "job", job)
		cmd := exec.Command(path.Join(cwd, "cmd", "run-analysis.sh"),
			strconv.FormatInt(int64(job.QueryPackId), 10),
			job.QueryLanguage, job.DBOwner, job.DBRepo)

		out, err := cmd.CombinedOutput()
		if err != nil {
			slog.Error("Analysis command failed: exit code: ", "error", err, "job", job)
			slog.Error("Analysis command failed: ", "job", job, "output", out)
			continue
		}
		slog.Debug("Analysis: finished", "job", job)

		// Get the SARIF ouput location
		sr := bufio.NewScanner(bytes.NewReader(out))
		sr.Split(bufio.ScanLines)
		for {
			more := sr.Scan()
			if !more {
				slog.Error("RunJob: analysis command failed to report result: ", "output", out)
				break
			}
			fields := strings.Fields(sr.Text())
			if len(fields) >= 3 {
				if fields[0] == "run-analysis-output" {
					slog.Debug("Analysis finished: ", "job", job, "location", fields[2])
					results <- AnalyzeResult{fields[2]}
					break
				}
			}
		}
	}
}

func init() {
	Jobs = make(chan AnalyzeJob, 10)
	Results = make(chan AnalyzeResult, 10)
	NumWorkers = 2

	for id := 1; id <= NumWorkers; id++ {
		go worker(id, Jobs, Results)
	}
}
