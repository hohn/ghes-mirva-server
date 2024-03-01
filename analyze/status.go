package analyze

import (
	"log/slog"
	"net/http"

	"github.com/hohn/ghes-mirva-server/api"
)

func StatusResponse(w http.ResponseWriter) {
	st := new(api.StatusResponse)

	slog.Debug("Submitting status response", "session", st.SessionId)
}
