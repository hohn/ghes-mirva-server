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

	"github.com/hohn/ghes-mirva-server/common"
	co "github.com/hohn/ghes-mirva-server/common"
	"github.com/hohn/ghes-mirva-server/store"
)

var (
	NumWorkers int
	Jobs       chan co.AnalyzeJob
	Results    chan common.AnalyzeResult
)

func worker(wid int, jobs <-chan co.AnalyzeJob, results chan<- common.AnalyzeResult) {
	var job co.AnalyzeJob
	for {
		job = <-jobs
		slog.Debug("Picking up job", "job", job, "worker", wid)

		cwd, err := os.Getwd()
		if err != nil {
			slog.Error("RunJob: cwd problem: ", "error", err)
			continue
		}

		slog.Debug("Analysis: running", "job", job)
		store.SetStatus(job.MirvaRequestID, job.ORL, common.StatusQueued)
		cmd := exec.Command(path.Join(cwd, "cmd", "run-analysis.sh"),
			strconv.FormatInt(int64(job.QueryPackId), 10),
			job.QueryLanguage, job.ORL.Owner, job.ORL.Repo)

		out, err := cmd.CombinedOutput()
		if err != nil {
			slog.Error("Analysis command failed: exit code: ", "error", err, "job", job)
			slog.Error("Analysis command failed: ", "job", job, "output", out)
			store.SetStatus(job.MirvaRequestID, job.ORL, common.StatusError)
			continue
		}
		slog.Debug("Analysis run finished", "job", job)

		// Get the SARIF ouput location
		sr := bufio.NewScanner(bytes.NewReader(out))
		sr.Split(bufio.ScanLines)
		for {
			more := sr.Scan()
			if !more {
				slog.Error("Analysis run failed to report result: ", "output", out)
				break
			}
			fields := strings.Fields(sr.Text())
			if len(fields) >= 3 {
				if fields[0] == "run-analysis-output" {
					slog.Debug("Analysis run successful: ", "job", job, "location", fields[2])
					res := common.AnalyzeResult{
						RunAnalysisOutput: fields[2]}
					results <- res
					store.SetStatus(job.MirvaRequestID, job.ORL, common.StatusSuccess)
					store.SetResult(job.MirvaRequestID, job.ORL, res)
					break
				}
			}
		}
	}
}

func init() {
	Jobs = make(chan co.AnalyzeJob, 10)
	Results = make(chan common.AnalyzeResult, 10)
	NumWorkers = 2

	for id := 1; id <= NumWorkers; id++ {
		go worker(id, Jobs, Results)
	}
}
