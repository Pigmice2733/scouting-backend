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

	fmt.Println("Initialized router...")
}

func (s *Server) initializeDB(dbFileName string) {
	var err error
	s.DB, err = sql.Open("sqlite3", dbFileName)
	if err != nil {
		log.Fatal(err)
	}

	const eventTableCreationQuery = `
	CREATE TABLE IF NOT EXISTS events
	(
		id INTEGER PRIMARY KEY,
		key TEXT NOT NULL,
		name TEXT NOT NULL,
		date TEXT NOT NULL
	)`

	const matchTableCreationQuery = `
	CREATE TABLE IF NOT EXISTS matches
	(
		id INTEGER PRIMARY KEY,
		eventID INT NOT NULL,
		winningAlliance TEXT,
		FOREIGN KEY(eventID) REFERENCES events(id)
	)`

	if _, err := s.DB.Exec(eventTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := s.DB.Exec(matchTableCreationQuery); err != nil {
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
	var e event
	err = getEvent(s.DB, id, &e)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	matches, err := getMatches(s.DB, e)

	fullE := &fullEvent{
		Event:   e,
		Matches: matches,
	}

	respondWithJSON(w, http.StatusOK, fullE)
}

func (s *Server) createEvents(events []event) error {
	for _, event := range events {
		err := createEvent(s.DB, event)
		if err != nil {
			fmt.Printf(fmt.Sprintf("Error processing TBA data '%s' in data '%v'", err.Error(), event))
			return err
		}
	}
	return nil
}
