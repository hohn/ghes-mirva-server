// Storage for job related information.  There is only one storage unit,
// hence no exported structs with member functions.
package store

import (
	"log/slog"
	"sync"

	"github.com/hohn/ghes-mirva-server/api"
	co "github.com/hohn/ghes-mirva-server/common"
)

var (
	info   map[co.JobSpec]co.JobInfo
	status map[co.JobSpec]co.Status
	result map[co.JobSpec]co.AnalyzeResult
	mutex  sync.Mutex
)

func SetResult(sessionid int, orl co.OwnerRepo, ar co.AnalyzeResult) {
	mutex.Lock()
	defer mutex.Unlock()
	result[co.JobSpec{sessionid, orl}] = ar
}

func SetStatus(sessionid int, orl co.OwnerRepo, s co.Status) {
	mutex.Lock()
	defer mutex.Unlock()
	status[co.JobSpec{sessionid, orl}] = s
}

func GetStatus(sessionid int, orl co.OwnerRepo) co.Status {
	mutex.Lock()
	defer mutex.Unlock()
	return status[co.JobSpec{sessionid, orl}]
}

func GetJobInfo(js co.JobSpec) co.JobInfo {
	mutex.Lock()
	defer mutex.Unlock()
	return info[js]
}

func StatusResponse() {
	st := new(api.StatusResponse)

	slog.Debug("Submitting status response", "session", st.SessionId)
}
