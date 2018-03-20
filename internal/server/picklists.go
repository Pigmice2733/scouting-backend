package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Pigmice2733/scouting-backend/internal/store/picklist"

	"github.com/Pigmice2733/scouting-backend/internal/respond"
	"github.com/Pigmice2733/scouting-backend/internal/store"

	"github.com/gorilla/mux"
)

func (s *Server) picklistHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	p, err := s.store.Picklist.Get(id)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			s.logger.LogRequestError(r, fmt.Errorf("getting picklist %d: %v", id, err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	respond.JSON(w, p)
}

func (s *Server) picklistsHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(keyUsernameCtx).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	bPicklists, err := s.store.Picklist.GetBasicPicklists(username)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("getting picklists: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respond.JSON(w, bPicklists)
}

func (s *Server) newPicklistHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(keyUsernameCtx).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var p picklist.Picklist
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	p.Owner = username

	id, err := s.store.Picklist.Insert(p)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("submitting picklist: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respond.JSON(w, id)
}

func (s *Server) updatePicklistHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(keyUsernameCtx).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var p picklist.Picklist
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	p.Owner = username
	p.ID = id

	realOwner, err := s.store.Picklist.GetOwner(p.ID)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			s.logger.LogRequestError(r, fmt.Errorf("getting picklist owner: %v", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if realOwner != p.Owner {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err := s.store.Picklist.Update(p); err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("submitting picklist: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) picklistEventHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(keyUsernameCtx).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	eventKey := mux.Vars(r)["eventKey"]

	bPicklists, err := s.store.Picklist.GetByEvent(username, eventKey)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("getting picklists: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respond.JSON(w, bPicklists)
}
