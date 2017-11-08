// server.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// A Server is an instance of the scouting server
type Server struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize setups up the database and router for the server
func (s *Server) Initialize(dbFileName string) {
	s.initializeDB(dbFileName)
	s.initializeRouter()
}

// Run starts a running the server on the specified address
func (s *Server) Run(addr string) {
	fmt.Println("Server up and running!")
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
		fmt.Printf("TBA polling failed with error %s\n", err)
		return
	}

	req.Header.Set("X-TBA-Auth-Key", apikey)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		fmt.Printf("TBA polling request failed with error %s\n", err)
		return
	}
	eventData, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(eventData, &tbaEvents)

	s.clearEventTable()

	var events []event

	for _, tbaEvent := range tbaEvents {
		date, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			fmt.Println("Error in processing TBA time data " + err.Error())
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

	fmt.Println("Polled TBA...")
}

func (s *Server) initializeRouter() {
	s.Router = mux.NewRouter().StrictSlash(true)

	s.Router.Handle("/events", wrapHandler(s.getEvents, "getEvents")).Methods("GET")
	s.Router.Handle("/events/{id:[0-9]+}", wrapHandler(s.getEvent, "getEvent")).Methods("GET")
	s.Router.Handle("/events/{eventID:[0-9]+}/{matchID:[0-9]+}", wrapHandler(s.getMatch, "getMatch")).Methods("GET")
	s.Router.Handle("/events/{eventID:[0-9]+}/{matchID:[0-9]+}", wrapHandler(s.postReport, "postReport")).Methods("POST")
	s.Router.Handle("/events/{eventID:[0-9]+}/{matchID:[0-9]+}/{team:[0-9]+}", wrapHandler(s.updateReport, "putReport")).Methods("PUT")

	fmt.Println("Initialized router...")
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

	fmt.Println("Initialized database...")
}

func wrapHandler(inner http.HandlerFunc, name string) http.Handler {
	return gziphandler.GzipHandler(Logger(inner, name))
}

func (s *Server) clearEventTable() {
	s.DB.Exec("DELETE FROM events")
	s.DB.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM events) WHERE name=\"events\"")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", "Accept-Encoding")

	w.Header().Set("Cache-Control", "public, max-age=60")

	w.WriteHeader(code)
	w.Write(response)
}

func (s *Server) getEvents(w http.ResponseWriter, r *http.Request) {
	events, err := getEvents(s.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, events)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}
	e := event{
		ID: id,
	}
	err = e.getEvent(s.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Non-existent event ID")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	matches, err := getMatches(s.DB, e)

	if err != nil && err != sql.ErrNoRows {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	fullEvent := &fullEvent{
		Event:   e,
		Matches: matches,
	}

	respondWithJSON(w, http.StatusOK, fullEvent)
}

func (s *Server) getMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.Atoi(vars["eventID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}
	matchID, err := strconv.Atoi(vars["matchID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}
	e := event{
		ID: eventID,
	}
	err = e.getEvent(s.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Non-existent event ID")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	partialMatch := match{
		ID:      matchID,
		EventID: eventID,
	}
	err = partialMatch.getMatch(s.DB)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Non-existent match ID under event ID")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
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
			respondWithError(w, http.StatusInternalServerError, err.Error())
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
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	fullMatch := &fullMatch{
		Match:        partialMatch,
		RedAlliance:  redAlliance,
		BlueAlliance: blueAlliance,
	}

	respondWithJSON(w, http.StatusOK, fullMatch)
}

func (s *Server) postReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID, err := strconv.Atoi(vars["matchID"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	var report reportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchID, report)

	if err != nil {
		if err != sql.ErrNoRows {
			respondWithError(w, http.StatusInternalServerError, err.Error())
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
			respondWithError(w, http.StatusBadRequest, "More than three reports for a single alliance not permitted")
			return
		}

		if teamReportAlreadyExists {
			respondWithError(w, http.StatusConflict, "Report post for team already exists, use 'PUT' to update")
		}

		a.updateAlliance(s.DB)
	}

	if err := report.createReport(s.DB, allianceID); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, report)
}

func (s *Server) updateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	matchID, err := strconv.Atoi(vars["matchID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	teamNumber, err := strconv.Atoi(vars["team"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid team number")
		return
	}

	var report reportData
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&report); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	allianceID, a, err := s.findAlliance(matchID, report)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Report does not exist, use 'POST' to create a report")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	teamReportExists := (a.Team1 == report.Team || a.Team2 == report.Team || a.Team3 == report.Team)

	if !teamReportExists {
		respondWithError(w, http.StatusNotFound, "Report does not exist, use 'POST' to create a report")
	}

	if report.Team != teamNumber {
		respondWithError(w, http.StatusBadRequest, "Report team does not match team specified in URI")
		return
	}

	if err := report.updateReport(s.DB, allianceID); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, report)
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
			fmt.Printf(fmt.Sprintf("Error processing TBA data '%s' in data '%v'", err.Error(), event))
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
