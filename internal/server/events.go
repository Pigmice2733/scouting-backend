package server

import (
	"fmt"
	"net/http"

	"github.com/Pigmice2733/scouting-backend/internal/store/event"

	"github.com/Pigmice2733/scouting-backend/internal/respond"
	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/gorilla/mux"
)

func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=86400") // 24 hour max age

	bEvents, err := s.store.Event.GetBasicEvents()
	if err == store.ErrNoResults {
		bEvents = []event.BasicEvent{}
	} else if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("getting basic events: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respond.JSON(w, bEvents)
}

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	eventKey := mux.Vars(r)["eventKey"]

	event, err := s.store.Event.Get(eventKey, s.store.Match)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			s.logger.LogRequestError(r, fmt.Errorf("getting event: %v", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	respond.JSON(w, event)
}

func (s *Server) matchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventKey, matchKey := vars["eventKey"], vars["matchKey"]

	match, err := s.store.Match.Get(eventKey, matchKey, s.store.Alliance)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			s.logger.LogRequestError(r, fmt.Errorf("getting match: %v", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	respond.JSON(w, match)
}
