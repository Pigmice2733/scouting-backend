// data_models_test.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

var s Server

func TestMain(m *testing.M) {
	s = Server{}
	s.Initialize("testing.db")

	exitCode := m.Run()

	os.Exit(exitCode)
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

func ensureTableExists(creationQuery string) {
	if _, err := s.DB.Exec(creationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable(t string) {
	delete := fmt.Sprintf("DELETE FROM %s", t)
	s.DB.Exec(delete)
	resetID := fmt.Sprintf("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM %s) WHERE name=\"%s\"", t, t)
	s.DB.Exec(resetID)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addEvents(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		s.DB.Exec("INSERT INTO events (key, name, date) VALUES (?, ?, ?)", (strconv.Itoa(i + 1)), ("Event " + strconv.Itoa(i+1)), time.Now().UTC().Format(time.RFC3339))
	}
}

func addMatches(count int, eID int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		s.DB.Exec("INSERT INTO matches (eventID) VALUES (?)", eID)
	}
}

func TestEmptyEventTable(t *testing.T) {
	ensureTableExists(eventTableCreationQuery)
	clearTable("events")

	req, _ := http.NewRequest("GET", "/events", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetEventMultiple(t *testing.T) {
	ensureTableExists(eventTableCreationQuery)
	clearTable("events")

	addEvents(2)

	req, _ := http.NewRequest("GET", "/events", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	d := []event{}

	json.Unmarshal(response.Body.Bytes(), &d)

	if d[0].Name != "Event 1" {
		t.Errorf("Expected event name to be 'Event 1'. Got '%v' instead.", d[0].Name)
	}
}

func TestGetMatchMultiple(t *testing.T) {
	ensureTableExists(eventTableCreationQuery)
	ensureTableExists(matchTableCreationQuery)
	clearTable("events")
	clearTable("matches")

	addEvents(1)
	addMatches(5, 1)

	req, _ := http.NewRequest("GET", "/events/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	fe := fullEvent{}

	json.Unmarshal(response.Body.Bytes(), &fe)

	if fe.Matches[1].EventID != 1 {
		t.Errorf("Expected event ID to be '1'. Got '%v' instead.", fe.Matches[0].EventID)
	}
}
