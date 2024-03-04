package analyze

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/hohn/ghes-mirva-server/api"
	co "github.com/hohn/ghes-mirva-server/common"
)

func StatusResponse(w http.ResponseWriter, js co.JobSpec, ji co.JobInfo) {
	st := new(api.StatusResponse)
	slog.Debug("Submitting status response", "session", st.SessionId)

	scanned := api.ScannedRepo{
		Repository:        api.Repository{},
		AnalysisStatus:    ji.Status.StatusString(),
		ResultCount:       0, // FIXME
		ArtifactSizeBytes: 0, // FIXME
	}

	all_scanned := []api.ScannedRepo{scanned}

	status := api.StatusResponse{
		SessionId:            js.ID,
		ControllerRepo:       api.ControllerRepo{},
		Actor:                api.Actor{},
		QueryLanguage:        ji.QueryLanguage,
		QueryPackURL:         "",
		CreatedAt:            ji.CreatedAt,
		UpdatedAt:            ji.UpdatedAt,
		ActionsWorkflowRunID: 0,
		Status:               ji.Status.StatusString(),
		ScannedRepositories:  all_scanned,
		SkippedRepositories:  ji.SkippedRepositories,
	}

	// Encode the response as JSON
	submit_response, err := json.Marshal(status)
	if err != nil {
		slog.Error("Error encoding response as JSON:",
			"error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send analysisReposJSON via ResponseWriter
	w.Header().Set("Content-Type", "application/json")
	w.Write(submit_response)
}
