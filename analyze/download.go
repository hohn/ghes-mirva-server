package analyze

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hohn/ghes-mirva-server/api"
	co "github.com/hohn/ghes-mirva-server/common"
	"github.com/hohn/ghes-mirva-server/store"
)

func DownloadResponse(w http.ResponseWriter, js co.JobSpec, vaid int) {
	slog.Debug("Forming download response", "session", vaid, "job", js)

	astat := store.GetStatus(vaid, js.OwnerRepo)

	var dlr api.DownloadResponse
	if astat == co.StatusSuccess {
		// We're looking for paths like
		// codeql/sarif/google/flatbuffers/google_flatbuffers.sarif
		ar := store.GetResult(js)

		//	FIXME ar.RunAnalysisOutput: "/Users/hohn/local/ghes-mirva-server/var/codeql/sarif/localrun/psycopg/psycopg2/psycopg_psycopg2.sarif"
		// hostname, err := os.Hostname()
		// if err != nil {
		// 	slog.Error("No host name found")
		// 	return
		// }
		hostname := "localhost" // FIXME

		au := fmt.Sprintf(
			"http://%s:8080/download-server/%s", hostname, ar.RunAnalysisOutput)

		dlr = api.DownloadResponse{
			Repository: api.DownloadRepo{
				Name:     js.Repo,
				FullName: fmt.Sprintf("%s/%s", js.Owner, js.Repo),
			},
			AnalysisStatus:       astat.ToExternalString(),
			ResultCount:          123, // FIXME
			ArtifactSizeBytes:    123, // FIXME
			DatabaseCommitSha:    "do-we-use-dcs-p",
			SourceLocationPrefix: "do-we-use-slp-p",
			// FIXME 404 page not found
			ArtifactURL: au,
		}
	} else {
		dlr = api.DownloadResponse{
			Repository: api.DownloadRepo{
				Name:     js.Repo,
				FullName: fmt.Sprintf("%s/%s", js.Owner, js.Repo),
			},
			AnalysisStatus:       astat.ToExternalString(),
			ResultCount:          0,
			ArtifactSizeBytes:    0,
			DatabaseCommitSha:    "",
			SourceLocationPrefix: "/not/relevant/here",
			ArtifactURL:          "",
		}
	}

	// Encode the response as JSON
	jdlr, err := json.Marshal(dlr)
	if err != nil {
		slog.Error("Error encoding response as JSON:",
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send analysisReposJSON via ResponseWriter
	w.Header().Set("Content-Type", "application/json")
	w.Write(jdlr)

}
func FileDownload(w http.ResponseWriter, path string) {
	slog.Debug("Readying file for upload", "path", path)

	fpath := path
	if !filepath.IsAbs(path) {
		fpath = "/" + path
	}

	file, err := os.Open(fpath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set headers
	fname := filepath.Base(fpath)
	w.Header().Set("Content-Disposition", "attachment; filename="+fname)
	w.Header().Set("Content-Type", "application/octet-stream")

	// Copy the file contents to the response writer
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to send file", http.StatusInternalServerError)
		return
	}

	slog.Debug("Uploaded file", "path", fpath)

}
