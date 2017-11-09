// server.go

package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/logger"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/sha3"
)

// A Server is an instance of the scouting server
type Server struct {
	Router *mux.Router
	DB     *sql.DB
	logger logger.Service
}

// New creates a new server given a db file and a io.Writer for logging
func New(dbFileName string, logWriter io.Writer, environment string) *Server {
	s := &Server{}

	s.logger = logger.New(logWriter, logger.Settings{
		Debug: environment == "dev",
		Info:  true,
		Error: true,
	})

	s.initializeDB(dbFileName)
	s.initializeRouter()

	return s
}

// Run starts a running the server on the specified address
func (s *Server) Run(addr string) {
	s.logger.Infof("Server up and running!")
	log.Fatal(http.ListenAndServe(addr, s.Router))
}

// PollTBA polls The Blue Alliance api for scouting data
func (s *Server) PollTBA(year string, apikey string) {

	tbaAddress := "https://www.thebluealliance.com/api/v3"

	type tbaEvent struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Date string `json:"start_date"`
	}
	var tbaEvents []tbaEvent

	req, err := http.NewRequest("GET", tbaAddress+"/events/"+year, nil)
	if err != nil {
		s.logger.Errorf("TBA polling failed with error %s\n", err)
		return
	}

	req.Header.Set("X-TBA-Auth-Key", apikey)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		s.logger.Errorf("TBA polling request failed with error %s\n", err)
		return
	}
	eventData, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(eventData, &tbaEvents)

	s.clearEventTable()

	var events []event

	for _, tbaEvent := range tbaEvents {
		date, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			s.logger.Errorf("Error in processing TBA time data " + err.Error())
			continue
		}
		newEvent := event{
			Key:  tbaEvent.Key,
			Name: tbaEvent.Name,
			Date: date,
		}
		events = append(events, newEvent)
	}

	s.createEvents(events)

	s.logger.Infof("Polled TBA...")
}

func (s *Server) initializeRouter() {
	s.Router = mux.NewRouter().StrictSlash(true)

	s.Router.Handle("/events", wrapHandler(s.getEvents, "getEvents", s.logger)).Methods("GET")
	s.Router.Handle("/events/{id:[0-9]+}", wrapHandler(s.getEvent, "getEvent", s.logger)).Methods("GET")
	s.Router.Handle("/events/{eventID:[0-9]+}/{matchID:[0-9]+}", wrapHandler(s.getMatch, "getMatch", s.logger)).Methods("GET")
	s.Router.Handle("/events/{eventID:[0-9]+}/{matchID:[0-9]+}", wrapHandler(s.postReport, "postReport", s.logger)).Methods("POST")
	s.Router.Handle("/events/{eventID:[0-9]+}/{matchID:[0-9]+}/{team:[0-9]+}", wrapHandler(s.updateReport, "putReport", s.logger)).Methods("PUT")

	s.logger.Infof("Initialized router...")
}

func (s *Server) initializeDB(dbFileName string) {
	var err error
	s.DB, err = sql.Open("sqlite3", dbFileName)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := s.DB.Exec(eventTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := s.DB.Exec(matchTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := s.DB.Exec(allianceTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	s.logger.Infof("Initialized database...")
}

func wrapHandler(inner http.HandlerFunc, name string, logger logger.Service) http.Handler {
	return gziphandler.GzipHandler(logger.Middleware(inner, name))
}

func (s *Server) clearEventTable() {
	s.DB.Exec("DELETE FROM events")
	s.DB.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM events) WHERE name=\"events\"")
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
	events, err := getEvents(s.DB)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ifNoneMatch := r.Header["If-None-Match"]
	// Cache for 24 hours = 1440 minutes
	respondGetJSON(w, http.StatusOK, events, 1440, ifNoneMatch)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}
	e := event{
		ID: id,
	}
	err = e.getEvent(s.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Non-existent event ID")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	matches, err := getMatches(s.DB, e)

	if err != nil && err != sql.ErrNoRows {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	fullEvent := &fullEvent{
		Event:   e,
		Matches: matches,
	}

	ifNoneMatch := r.Header["If-None-Match"]
	// Cache for 1 minute
	respondGetJSON(w, http.StatusOK, fullEvent, 1, ifNoneMatch)
}

func (s *Server) getMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.Atoi(vars["eventID"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}
	matchID, err := strconv.Atoi(vars["matchID"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}
	e := event{
		ID: eventID,
	}
	err = e.getEvent(s.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Non-existent event ID")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	partialMatch := match{
		ID:      matchID,
		EventID: eventID,
	}
	err = partialMatch.getMatch(s.DB)

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Non-existent match ID under event ID")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	blueAlliance := alliance{
		MatchID: matchID,
		IsBlue:  true,
	}
	redAlliance := alliance{
		MatchID: matchID,
		IsBlue:  false,
	}

	_, err = blueAlliance.getAlliance(s.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			blueAlliance = alliance{
				MatchID: matchID,
				IsBlue:  true,
			}
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	_, err = redAlliance.getAlliance(s.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			redAlliance = alliance{
				MatchID: matchID,
				IsBlue:  false,
			}
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	fullMatch := &fullMatch{
		Match:        partialMatch,
		RedAlliance:  redAlliance,
		BlueAlliance: blueAlliance,
	}

	ifNoneMatch := r.Header["If-None-Match"]
	// Cache for 1 minutes
	respondGetJSON(w, http.StatusOK, fullMatch, 1, ifNoneMatch)
}

func (s *Server) postReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID, err := strconv.Atoi(vars["matchID"])

	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	var report reportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchID, report)

	if err != nil {
		if err != sql.ErrNoRows {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.Team1 = report.Team
		a.createAlliance(s.DB)
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

		a.updateAlliance(s.DB)
	}

	if err := report.createReport(s.DB, allianceID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondSetJSON(w, http.StatusCreated, report)
}

func (s *Server) updateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	matchID, err := strconv.Atoi(vars["matchID"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	teamNumber, err := strconv.Atoi(vars["team"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid team number")
		return
	}

	var report reportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchID, report)

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Report does not exist, use 'POST' to create a report")
		} else {
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

	if err := report.updateReport(s.DB, allianceID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondSetJSON(w, http.StatusOK, report)
}

func (s *Server) findAlliance(matchID int, report reportData) (int, alliance, error) {
	isBlue := true
	if report.Alliance == "red" {
		isBlue = false
	}

	a := alliance{
		MatchID: matchID,
		IsBlue:  isBlue,
		Score:   report.Score,
	}

	id, err := a.getAlliance(s.DB)
	return id, a, err
}

func (s *Server) createEvents(events []event) error {
	for _, event := range events {
		err := event.createEvent(s.DB)
		if err != nil {
			s.logger.Errorf("Error processing TBA data '%s' in data '%v'", err.Error(), event)
			return err
		}
	}
	return nil
}

const eventTableCreationQuery = `
CREATE TABLE IF NOT EXISTS events
(
	id   INTEGER PRIMARY KEY,
	key  TEXT    NOT NULL,
    name TEXT    NOT NULL,
    date TEXT    NOT NULL
)`

const matchTableCreationQuery = `
CREATE TABLE IF NOT EXISTS matches
(
	id              INTEGER PRIMARY KEY,
	eventID         INT     NOT NULL,
	winningAlliance TEXT,
	FOREIGN KEY(eventID) REFERENCES events(id)
)`

const allianceTableCreationQuery = `
CREATE TABLE IF NOT EXISTS alliances
(
	id      INTEGER PRIMARY KEY,
	matchID INT     NOT NULL,
	score   INT     NOT NULL,
	team1   INT,
	team2   INT,
	team3   INT,
	isBlue  INT     NOT NULL,
	FOREIGN KEY(matchID) REFERENCES matches(id)
)
`

const reportTableCreationQuery = `
CREATE TABLE IF NOT EXISTS reports
(
	id            INTEGER PRIMARY KEY,
	allianceID    INT     NOT NULL,
	teamNumber    INT     NOT NULL,
	score         INT     NOT NULL,
	crossedLine   INT,
	deliveredGear INT,
	autoFuel      INT,
    climbed       INT,
	gears         INT,
	teleopFuel    INT,
	FOREIGN KEY(allianceID) REFERENCES alliances(id)
)
`
