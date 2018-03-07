package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Pigmice2733/scouting-backend/internal/analysis"
	"github.com/Pigmice2733/scouting-backend/internal/store/report"
	"github.com/gorilla/mux"
)

func (s *Server) reportHandler(w http.ResponseWriter, r *http.Request) {
	var rep report.Report
	if err := json.NewDecoder(r.Body).Decode(&rep); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	rep.EventKey, rep.MatchKey = vars["eventKey"], vars["matchKey"]

	rep.Reporter = ""
	if reporter, ok := r.Context().Value(keyUsernameCtx).(string); ok {
		rep.Reporter = reporter
	}

	if !analysis.CompliantData(s.schema, rep.Stats) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := s.store.Report.Upsert(rep, s.store.Alliance); err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("upserting report: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
