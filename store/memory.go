// Storage for job related information.  There is only one storage unit,
// hence no exported structs with member functions.
package store

import (
	"log/slog"
	"sync"

	"github.com/hohn/ghes-mirva-server/api"
	"github.com/hohn/ghes-mirva-server/common"
	co "github.com/hohn/ghes-mirva-server/common"
)

type Status int

const (
	StatusInProgress = iota
	StatusQueued
	StatusError
	StatusSuccess
	StatusFailed
)

type JobSpec struct {
	id  int
	orl co.OwnerRepoLoc
}

var (
	status map[JobSpec]Status
	result map[JobSpec]common.AnalyzeResult
	mutex  sync.Mutex
)

func SetResult(sessionid int, orl co.OwnerRepoLoc, ar co.AnalyzeResult) {
	mutex.Lock()
	defer mutex.Unlock()
	result[JobSpec{sessionid, orl}] = ar
}

func SetStatus(sessionid int, orl co.OwnerRepoLoc, s Status) {
	mutex.Lock()
	defer mutex.Unlock()
	status[JobSpec{sessionid, orl}] = s
}

func GetStatus(sessionid int, orl co.OwnerRepoLoc) Status {
	mutex.Lock()
	defer mutex.Unlock()
	return status[JobSpec{sessionid, orl}]
}

func StatusResponse() {
	st := new(api.StatusResponse)

	slog.Debug("Submitting status response", "session", st.SessionId)
}
