package server

import (
	"fmt"
	"net/http"

	"github.com/Pigmice2733/scouting-backend/internal/analysis"
	"github.com/Pigmice2733/scouting-backend/internal/respond"
	"github.com/Pigmice2733/scouting-backend/internal/server/logic"
	"github.com/gorilla/mux"
)

func (s *Server) schemaHandler(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, s.schema)
}

func (s *Server) eventAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	eventKey := mux.Vars(r)["eventKey"]

	resp, err := logic.EventAnalysis(eventKey, s.schema, s.store.Report)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("analyzing event: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respond.JSON(w, resp)
}

func (s *Server) teamAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventKey, team := vars["eventKey"], vars["team"]

	resp, err := logic.Analyze(eventKey, []string{team}, s.schema, s.store.Report)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("analyzing event: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(resp) == 0 {
		respond.JSON(w, analysis.Results{})
		return
	}

	respond.JSON(w, resp[0].Stats)
}

func (s *Server) allianceAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventKey, matchKey, color := vars["eventKey"], vars["matchKey"], vars["color"]

	resp, err := logic.AllianceAnalysis(eventKey, matchKey, color, s.schema, s.store.Report)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("analyzing event: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respond.JSON(w, resp)
}
