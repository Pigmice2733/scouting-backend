package server

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/logger"
	"github.com/Pigmice2733/scouting-backend/store"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/sha3"
)

// A Server is an instance of the scouting server
type Server struct {
	Router    *mux.Router
	store     store.Service
	logger    logger.Service
	tbaAPIKey string
}

// New creates a new server given a db file and a io.Writer for logging
func New(store store.Service, logWriter io.Writer, tbaAPIKey string, environment string) *Server {
	s := &Server{}

	s.logger = logger.New(logWriter, logger.Settings{
		Debug: environment == "dev",
		Info:  true,
		Error: true,
	})

	s.store = store

	s.initializeRouter()

	s.logger = logger.New(logWriter, logger.Settings{
		Debug: environment == "dev",
		Info:  true,
		Error: true,
	})

	s.tbaAPIKey = tbaAPIKey

	return s
}

// Run starts a running the server on the specified address
func (s *Server) Run(addr string) {
	s.logger.Infof("Server up and running!")
	log.Fatal(http.ListenAndServe(addr, s.Router))
}

// PollTBA polls The Blue Alliance api to populate database
func (s *Server) PollTBA(year string) {
	s.pollTBAEvents(s.logger, "https://www.thebluealliance.com/api/v3", s.tbaAPIKey, year)
}

func (s *Server) initializeRouter() {
	s.Router = mux.NewRouter().StrictSlash(true)

	s.Router.Handle("/events", wrapHandler(s.getEvents, "getEvents", s.logger)).Methods("GET")
	s.Router.Handle("/events/{eventKey}", wrapHandler(s.getEvent, "getEvent", s.logger)).Methods("GET")
	s.Router.Handle("/events/{eventKey}/{matchKey}", wrapHandler(s.getMatch, "getMatch", s.logger)).Methods("GET")
	s.Router.Handle("/events/{eventKey}/{matchKey}", wrapHandler(s.postReport, "postReport", s.logger)).Methods("POST")
	s.Router.Handle("/events/{eventKey}/{matchKey}/{team:[0-9]+}", wrapHandler(s.updateReport, "putReport", s.logger)).Methods("PUT")

	s.logger.Infof("Initialized router...")
}

func wrapHandler(inner http.HandlerFunc, name string, logger logger.Service) http.Handler {
	return gziphandler.GzipHandler(logger.Middleware(inner, name))
}

func generateEtag(content []byte) (string, error) {
	hash := make([]byte, 32)
	d := sha3.NewShake256()
	// Write the response into the hash
	d.Write(content)
	// Read 32 bytes of output from the hash into h.
	_, err := d.Read(hash)

	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}

func respond(w http.ResponseWriter, code int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", "Accept-Encoding")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(code)
	w.Write(payload)
}

func respondError(w http.ResponseWriter, code int, message string) {
	payload, _ := json.Marshal(map[string]string{"error": message})
	w.Header().Set("Cache-Control", "no-cache")
	respond(w, code, payload)
}

// Use for setter methods - POST, DELETE, etc.
func respondSetJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Cache-Control", "no-cache")
	respond(w, code, response)
}

// Use for getter methods - GET, HEAD
func respondGetJSON(w http.ResponseWriter, code int, payload interface{}, cacheMinutes int, ifNoneMatch []string) {
	response, _ := json.Marshal(payload)
	contentETag, _ := generateEtag(response)
	cacheControl := fmt.Sprintf("public, max-age=%d", (cacheMinutes * 60))

	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("ETag", contentETag)

	for _, eTag := range ifNoneMatch {
		if eTag == contentETag {
			respond(w, http.StatusNotModified, nil)
			return
		}
	}

	respond(w, code, response)
}

func (s *Server) getEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.store.GetEvents()
	if err != nil {
		s.logger.Debugf("Error in getEvents %s", err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ifNoneMatch := r.Header["If-None-Match"]
	// Cache for 24 hours = 1440 minutes
	respondGetJSON(w, http.StatusOK, events, 1440, ifNoneMatch)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventKey := vars["eventKey"]
	e := &store.Event{
		Key: eventKey,
	}
	err := s.store.GetEvent(e)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Non-existent event key")
			return
		}
		s.logger.Debugf("Error in getEvent %s", err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	matches, err := s.pollTBAMatches("https://www.thebluealliance.com/api/v3", s.tbaAPIKey, e.Key)
	if err != nil {
		s.logger.Infof(err.Error())
	}

	matches, err = s.store.GetMatches(*e)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Debugf("Error in getEvents %s", err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	fullEvent := &store.FullEvent{
		Key:     e.Key,
		Name:    e.Name,
		Date:    e.Date,
		Matches: matches,
	}

	ifNoneMatch := r.Header["If-None-Match"]
	// Cache for 1 minute
	respondGetJSON(w, http.StatusOK, fullEvent, 1, ifNoneMatch)
}

func (s *Server) getMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventKey := vars["eventKey"]
	matchKey := vars["matchKey"]
	e := &store.Event{
		Key: eventKey,
	}

	err := s.store.GetEvent(e)

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Non-existent event key")
			return
		}
		s.logger.Debugf("Error in getMatch %s", err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	partialMatch := &store.Match{
		Key:      matchKey,
		EventKey: eventKey,
	}
	err = s.store.GetMatch(partialMatch)

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Non-existent match key under event key")
			return
		}
		s.logger.Debugf("Error in getMatch %s", err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	blueAlliance := &store.Alliance{
		MatchKey: matchKey,
		IsBlue:   true,
	}
	redAlliance := &store.Alliance{
		MatchKey: matchKey,
		IsBlue:   false,
	}

	_, err = s.store.GetAlliance(blueAlliance)
	if err != nil && err != sql.ErrNoRows {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = s.store.GetAlliance(redAlliance)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Debugf("Error in getMatch %s", err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	fullMatch := &store.FullMatch{
		Key:             partialMatch.Key,
		EventKey:        partialMatch.EventKey,
		WinningAlliance: partialMatch.WinningAlliance,
		RedAlliance:     *redAlliance,
		BlueAlliance:    *blueAlliance,
	}

	ifNoneMatch := r.Header["If-None-Match"]
	// Cache for 1 minute
	respondGetJSON(w, http.StatusOK, fullMatch, 1, ifNoneMatch)
}

func (s *Server) postReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchKey := vars["matchKey"]

	var report store.ReportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchKey, report)

	if err != nil {
		if err != sql.ErrNoRows {
			s.logger.Debugf("Error in postReport %s", err.Error())
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.Team1 = report.Team
		allianceID, err = s.store.CreateAlliance(a)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		teamReportAlreadyExists := false
		switch {
		case a.Team2 == 0:
			if a.Team1 == report.Team {
				teamReportAlreadyExists = true
			}
			a.Team2 = report.Team
		case a.Team3 == 0:
			if a.Team1 == report.Team || a.Team2 == report.Team {
				teamReportAlreadyExists = true
			}
			a.Team3 = report.Team
		default:
			respondError(w, http.StatusBadRequest, "More than three reports for a single alliance not permitted")
			return
		}

		if teamReportAlreadyExists {
			respondError(w, http.StatusConflict, "Report post for team already exists, use 'PUT' to update")
		}

		err := s.store.UpdateAlliance(a)
		if err != nil {
			s.logger.Debugf("Error in postReport %s", err.Error())
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := s.store.CreateReport(report, allianceID); err != nil {
		s.logger.Debugf("Error in postReport %s", err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondSetJSON(w, http.StatusCreated, report)
}

func (s *Server) updateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchKey := vars["matchKey"]
	teamNumber, err := strconv.Atoi(vars["team"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid team number")
		return
	}

	var report store.ReportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchKey, report)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Report does not exist, use 'POST' to create a report")
		} else {
			s.logger.Debugf("Error in updateReport %s", err.Error())
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	teamReportExists := (a.Team1 == report.Team || a.Team2 == report.Team || a.Team3 == report.Team)

	if !teamReportExists {
		respondError(w, http.StatusNotFound, "Report does not exist, use 'POST' to create a report")
	}

	if report.Team != teamNumber {
		respondError(w, http.StatusBadRequest, "Report team does not match team specified in URI")
		return
	}

	if err := s.store.UpdateReport(report, allianceID); err != nil {
		s.logger.Debugf("Error in updateReport %v\n", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondSetJSON(w, http.StatusOK, report)
}

func (s *Server) findAlliance(matchKey string, report store.ReportData) (int, store.Alliance, error) {
	isBlue := true
	if report.Alliance == "red" {
		isBlue = false
	}

	a := &store.Alliance{
		MatchKey: matchKey,
		IsBlue:   isBlue,
		Score:    report.Score,
	}

	id, err := s.store.GetAlliance(a)
	return id, *a, err
}

func (s *Server) createEvents(events []store.Event) error {
	for _, event := range events {
		err := s.store.CreateEvent(event)
		if err != nil {
			s.logger.Errorf("Error processing TBA data '%s' in data '%v'", err.Error(), event)
			return err
		}
	}
	return nil
}
