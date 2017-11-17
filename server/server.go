package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/server/logger"
	"github.com/Pigmice2733/scouting-backend/server/store"
	"github.com/gorilla/mux"
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
func (s *Server) Run(addr string) error {
	s.logger.Infof("server up and running")
	return http.ListenAndServe(addr,
		s.logger.Middleware(
			gziphandler.GzipHandler(
				addDefaultHeaders(s.Router))))
}

// PollTBA polls The Blue Alliance api to populate database
func (s *Server) PollTBA(year string) error {
	return s.pollTBAEvents(s.logger, "https://www.thebluealliance.com/api/v3", s.tbaAPIKey, year)
}

func (s *Server) initializeRouter() {
	s.Router = mux.NewRouter().StrictSlash(true)

	s.Router.HandleFunc("/events", s.getEvents).Methods("GET")
	s.Router.HandleFunc("/events/{eventKey}", s.getEvent).Methods("GET")
	s.Router.HandleFunc("/events/{eventKey}/{matchKey}", s.getMatch).Methods("GET")
	s.Router.HandleFunc("/events/{eventKey}/{matchKey}", s.postReport).Methods("POST")
	s.Router.HandleFunc("/events/{eventKey}/{matchKey}/{team:[0-9]+}", s.updateReport).Methods("PUT")

	s.logger.Infof("initialized router...")
}

// REST Endpoint Handlers -----------------------

func (s *Server) getEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.store.GetEvents()
	if err != nil {
		s.logger.Errorf("error: getting events: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(events)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1440)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.Errorf("error: handling eTag %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if foundMatchingEtag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(response); err != nil {
		s.logger.Errorf("error: writing json []byte response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	eventKey := vars["eventKey"]
	e := &store.Event{
		Key: eventKey,
	}
	err := s.store.GetEvent(e)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, "non-existent event key", http.StatusNotFound)
			return
		}
		s.logger.Errorf("error: getting events: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	matches, err := s.pollTBAMatches("https://www.thebluealliance.com/api/v3", s.tbaAPIKey, e.Key)
	if err != nil {
		s.logger.Infof(err.Error())

		matches, err = s.store.GetMatches(*e)
		if err != nil && err != store.ErrNoResults {
			s.logger.Errorf("error: getting events: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	fullEvent := &store.FullEvent{
		Key:     e.Key,
		Name:    e.Name,
		Date:    e.Date,
		Matches: matches,
	}

	response, err := json.Marshal(fullEvent)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.Errorf("error: handling eTag %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if foundMatchingEtag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(response); err != nil {
		s.logger.Errorf("error: writing json []byte response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
		if err == store.ErrNoResults {
			http.Error(w, "non-existent event key", http.StatusNotFound)
			return
		}
		s.logger.Errorf("error: getting match: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	partialMatch := &store.Match{
		Key:      matchKey,
		EventKey: eventKey,
	}
	err = s.store.GetMatch(partialMatch)

	if err != nil {
		if err == store.ErrNoResults {
			message := fmt.Sprintf("non-existent match key '%v' under event key '%v'", matchKey, eventKey)
			http.Error(w, message, http.StatusNotFound)
			return
		}
		s.logger.Errorf("error: getting match: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting alliance: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_, err = s.store.GetAlliance(redAlliance)
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting alliance: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fullMatch := &store.FullMatch{
		Key:             partialMatch.Key,
		EventKey:        partialMatch.EventKey,
		WinningAlliance: partialMatch.WinningAlliance,
		RedAlliance:     *redAlliance,
		BlueAlliance:    *blueAlliance,
	}

	response, err := json.Marshal(fullMatch)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.Errorf("error: handling eTag %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if foundMatchingEtag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(response); err != nil {
		s.logger.Errorf("error: writing json []byte response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) postReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	matchKey := vars["matchKey"]

	var report store.ReportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchKey, report)

	if err != nil {
		if err != store.ErrNoResults {
			s.logger.Errorf("error: nothing present in response: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		a.Team1 = report.Team
		allianceID, err = s.store.CreateAlliance(a)
		if err != nil {
			http.Error(w, "no corresponding match for posted report", http.StatusBadRequest)
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
			http.Error(w, "more than three reports for a single alliance not permitted", http.StatusBadRequest)
			return
		}

		if teamReportAlreadyExists {
			http.Error(w, "report for team already exists, use 'PUT' to update", http.StatusConflict)
			return
		}

		err := s.store.UpdateAlliance(a)
		if err != nil {
			s.logger.Errorf("error: postreport: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if err := s.store.CreateReport(report, allianceID); err != nil {
		s.logger.Errorf("error: postreport: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(report)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(response); err != nil {
		s.logger.Errorf("error: writing json []byte response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) updateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	matchKey := vars["matchKey"]
	teamNumber, err := strconv.Atoi(vars["team"])
	if err != nil {
		http.Error(w, "invalid team number", http.StatusBadRequest)
		return
	}

	var report store.ReportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchKey, report)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, "report does not exist, use 'POST' to create a new report", http.StatusNotFound)
			return
		}
		s.logger.Errorf("error: updateReport %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	teamReportExists := (a.Team1 == report.Team || a.Team2 == report.Team || a.Team3 == report.Team)

	if !teamReportExists {
		http.Error(w, "report does not exist, use 'POST' to create a new report", http.StatusNotFound)
		return
	}

	if report.Team != teamNumber {
		http.Error(w, "report team does not match team specified in URI", http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateReport(report, allianceID); err != nil {
		s.logger.Errorf("error: updateReport %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(report)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(response); err != nil {
		s.logger.Errorf("error: writing json []byte response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
			s.logger.Errorf("error: processing TBA data '%v' in data '%v'\n", err, event)
			return err
		}
	}
	return nil
}

func normalCache(w http.ResponseWriter, cacheMinutes int) {
	cacheControl := fmt.Sprintf("public, max-age=%d", (cacheMinutes * 60))
	if cacheMinutes == 0 {
		cacheControl = "no-cache"
	}

	w.Header().Set("Cache-Control", cacheControl)
}

func noCache(w http.ResponseWriter, cacheMinutes int) {
	w.Header().Set("Cache-Control", "no-store")
}

func addETags(w http.ResponseWriter, r *http.Request, response []byte) (bool, error) {
	contentETag, err := generateEtag(response)
	if err != nil {
		return false, err
	}

	ifNoneMatch := r.Header["If-None-Match"]

	for _, eTag := range ifNoneMatch {
		if eTag == contentETag {
			return true, nil
		}
	}

	w.Header().Set("ETag", contentETag)

	return false, nil
}

func addDefaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func generateEtag(content []byte) (string, error) {
	hash := make([]byte, 32)
	d := sha3.NewShake256()
	// Write the response into the hash
	if _, err := d.Write(content); err != nil {
		return "", err
	}
	// Read 32 bytes of output from the hash into h.
	if _, err := d.Read(hash); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash), nil
}

func isError(statusCode int) bool {
	// HTTP error codes are status codes 4xx-5xx
	if statusCode >= 400 && statusCode < 600 {
		return true
	}
	return false
}
