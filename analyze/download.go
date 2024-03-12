package analyze

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
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

		hostname, err := os.Hostname()
		if err != nil {
			slog.Error("No host name found")
			return
		}

		zfpath, err := PackageResults(ar, js.OwnerRepo, vaid)
		if err != nil {
			slog.Error("Error packaging results:", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		au := fmt.Sprintf(
			"http://%s:8080/download-server/%s", hostname, zfpath)
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

func PackageResults(ar co.AnalyzeResult, owre co.OwnerRepo, vaid int) (zipPath string, e error) {
	slog.Debug("Readying zip file with .sarif/.bqrs", "analyze-result", ar)

	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("No working directory")
		panic(err)
	}

	// Ensure the output directory exists
	dirpath := path.Join(cwd, "var", "codeql", "localrun", "results")
	if err := os.MkdirAll(dirpath, 0755); err != nil {
		slog.Error("Unable to create results output directory",
			"dir", dirpath)
		return "", err
	}

	// Create a new zip file
	zpath := path.Join(dirpath, fmt.Sprintf("results-%s-%s-%d.zip", owre.Owner, owre.Repo, vaid))

	zfile, err := os.Create(zpath)
	if err != nil {
		return "", err
	}
	defer zfile.Close()

	// Create a new zip writer
	zwriter := zip.NewWriter(zfile)
	defer zwriter.Close()

	// Add each result file to the zip archive
	names := []([]string){{ar.RunAnalysisSARIF, "results.sarif"}}
	for _, fpath := range names {
		file, err := os.Open(fpath[0])
		if err != nil {
			return "", err
		}
		defer file.Close()

		// Create a new file in the zip archive with custom name
		// The client is very specific:
		// if zf.Name != "results.sarif" && zf.Name != "results.bqrs" { continue }

		zipEntry, err := zwriter.Create(fpath[1])
		if err != nil {
			return "", err
		}

		// Copy the contents of the file to the zip entry
		_, err = io.Copy(zipEntry, file)
		if err != nil {
			return "", err
		}
	}
	return zpath, nil
}

func FileDownload(w http.ResponseWriter, path string) {
	slog.Debug("Sending zip file with .sarif/.bqrs", "path", path)

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
