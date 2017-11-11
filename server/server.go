package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/logger"
	"github.com/Pigmice2733/scouting-backend/store"
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
	return http.ListenAndServe(addr, s.Router)
}

// PollTBA polls The Blue Alliance api to populate database
func (s *Server) PollTBA(year string) error {
	return s.pollTBAEvents(s.logger, "https://www.thebluealliance.com/api/v3", s.tbaAPIKey, year)
}

func (s *Server) initializeRouter() {
	s.Router = mux.NewRouter().StrictSlash(true)

	s.Router.Handle("/events", wrapHandler(s.getEvents, s.logger, 1440)).Methods("GET")
	s.Router.Handle("/events/{eventKey}", wrapHandler(s.getEvent, s.logger, 1)).Methods("GET")
	s.Router.Handle("/events/{eventKey}/{matchKey}", wrapHandler(s.getMatch, s.logger, 1)).Methods("GET")
	s.Router.Handle("/events/{eventKey}/{matchKey}", wrapHandler(s.postReport, s.logger, 1)).Methods("POST")
	s.Router.Handle("/events/{eventKey}/{matchKey}/{team:[0-9]+}", wrapHandler(s.updateReport, s.logger, 1)).Methods("PUT")

	s.logger.Infof("initialized router...")
}

func wrapHandler(inner func(io.ReadCloser, map[string]string) (interface{}, int), logger logger.Service, cacheMinutes int) http.Handler {
	return gziphandler.GzipHandler(logger.Middleware(requestHandler(logger, cacheMinutes, inner)))
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

func requestHandler(logger logger.Service, cacheMinutes int, inner func(io.ReadCloser, map[string]string) (interface{}, int)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)

		payload, status := inner(r.Body, vars)
		ifNoneMatch := r.Header["If-None-Match"]

		if isError(status) {
			http.Error(w, payload.(string), status)
			return
		}

		response, err := json.Marshal(payload)
		if err != nil {
			logger.Errorf("error: marshalling json response %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		contentETag, err := generateEtag(response)
		if err != nil {
			logger.Errorf("error: generating eTag %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		cacheControl := fmt.Sprintf("public, max-age=%d", (cacheMinutes * 60))
		if cacheMinutes == 0 {
			cacheControl = "no-cache"
		}

		for _, eTag := range ifNoneMatch {
			if eTag == contentETag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
		w.Header().Set("Cache-Control", cacheControl)
		w.Header().Set("ETag", contentETag)
		w.WriteHeader(status)
		if _, err := w.Write(response); err != nil {
			logger.Errorf("error: writing json []byte response %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (s *Server) getEvents(data io.ReadCloser, vars map[string]string) (interface{}, int) {
	events, err := s.store.GetEvents()
	if err != nil {
		s.logger.Errorf("error: getting events: %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}
	return events, http.StatusOK
}

func (s *Server) getEvent(data io.ReadCloser, vars map[string]string) (interface{}, int) {
	eventKey := vars["eventKey"]
	e := &store.Event{
		Key: eventKey,
	}
	err := s.store.GetEvent(e)
	if err != nil {
		if err == store.ErrNoResults {
			return "non-existent event key", http.StatusNotFound
		}
		s.logger.Errorf("error: getting events: %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}

	matches, err := s.pollTBAMatches("https://www.thebluealliance.com/api/v3", s.tbaAPIKey, e.Key)
	if err != nil {
		s.logger.Infof(err.Error())

		matches, err = s.store.GetMatches(*e)
		if err != nil && err != store.ErrNoResults {
			s.logger.Errorf("error: getting events: %v\n", err)
			return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
		}
	}

	fullEvent := &store.FullEvent{
		Key:     e.Key,
		Name:    e.Name,
		Date:    e.Date,
		Matches: matches,
	}

	return fullEvent, http.StatusOK
}

func (s *Server) getMatch(data io.ReadCloser, vars map[string]string) (interface{}, int) {
	eventKey := vars["eventKey"]
	matchKey := vars["matchKey"]
	e := &store.Event{
		Key: eventKey,
	}

	err := s.store.GetEvent(e)

	if err != nil {
		if err == store.ErrNoResults {
			return "non-existent event key", http.StatusNotFound
		}
		s.logger.Errorf("error: getting match: %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}

	partialMatch := &store.Match{
		Key:      matchKey,
		EventKey: eventKey,
	}
	err = s.store.GetMatch(partialMatch)

	if err != nil {
		if err == store.ErrNoResults {
			message := fmt.Sprintf("non-existent match key '%v' under event key '%v'", matchKey, eventKey)
			return message, http.StatusNotFound
		}
		s.logger.Errorf("error: getting match: %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
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
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}

	_, err = s.store.GetAlliance(redAlliance)
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting alliance: %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}

	fullMatch := &store.FullMatch{
		Key:             partialMatch.Key,
		EventKey:        partialMatch.EventKey,
		WinningAlliance: partialMatch.WinningAlliance,
		RedAlliance:     *redAlliance,
		BlueAlliance:    *blueAlliance,
	}

	return fullMatch, http.StatusOK
}

func (s *Server) postReport(data io.ReadCloser, vars map[string]string) (interface{}, int) {
	matchKey := vars["matchKey"]

	var report store.ReportData

	if err := json.NewDecoder(data).Decode(&report); err != nil {
		return "invalid request payload", http.StatusBadRequest
	}
	defer data.Close()

	allianceID, a, err := s.findAlliance(matchKey, report)

	if err != nil {
		if err != store.ErrNoResults {
			s.logger.Errorf("error: nothing present in response: %v\n", err)
			return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
		}
		a.Team1 = report.Team
		allianceID, err = s.store.CreateAlliance(a)
		if err != nil {
			return "no corresponding match for posted report", http.StatusBadRequest
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
			return "more than three reports for a single alliance not permitted", http.StatusBadRequest
		}

		if teamReportAlreadyExists {
			return "report for team already exists, use 'PUT' to update", http.StatusConflict
		}

		err := s.store.UpdateAlliance(a)
		if err != nil {
			s.logger.Errorf("error: postreport: %v\n", err)
			return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
		}
	}

	if err := s.store.CreateReport(report, allianceID); err != nil {
		s.logger.Errorf("error: postreport: %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}

	return report, http.StatusCreated
}

func (s *Server) updateReport(data io.ReadCloser, vars map[string]string) (interface{}, int) {
	matchKey := vars["matchKey"]
	teamNumber, err := strconv.Atoi(vars["team"])
	if err != nil {
		return "invalid team number", http.StatusBadRequest
	}

	var report store.ReportData
	decoder := json.NewDecoder(data)

	if err := decoder.Decode(&report); err != nil {
		return "invalid request payload", http.StatusBadRequest
	}
	defer data.Close()

	allianceID, a, err := s.findAlliance(matchKey, report)
	if err != nil {
		if err == store.ErrNoResults {
			return "report does not exist, use 'POST' to create a report", http.StatusNotFound
		}
		s.logger.Errorf("error: updateReport %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}

	teamReportExists := (a.Team1 == report.Team || a.Team2 == report.Team || a.Team3 == report.Team)

	if !teamReportExists {
		return "report does not exist, use 'POST' to create a report", http.StatusNotFound
	}

	if report.Team != teamNumber {
		return "report team does not match team specified in URI", http.StatusBadRequest
	}

	if err := s.store.UpdateReport(report, allianceID); err != nil {
		s.logger.Errorf("error: updateReport %v\n", err)
		return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
	}

	return report, http.StatusOK
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
