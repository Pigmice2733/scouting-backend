// data_models_test.go

package main

import (
	"bytes"
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

var s *Server

func TestMain(m *testing.M) {
	s = New("testing.db", os.Stdout, "dev")
	exitCode := m.Run()
	os.Exit(exitCode)
}

func setupTables() {
	ensureTableExists(eventTableCreationQuery)
	ensureTableExists(matchTableCreationQuery)
	ensureTableExists(allianceTableCreationQuery)
	ensureTableExists(reportTableCreationQuery)
	clearTable("events")
	clearTable("matches")
	clearTable("alliances")
	clearTable("reports")
}

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
		_, err := s.DB.Exec("INSERT INTO events (key, name, date) VALUES (?, ?, ?)", (strconv.Itoa(i + 1)), ("Event " + strconv.Itoa(i+1)), time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func addMatches(count int, eID int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		_, err := s.DB.Exec("INSERT INTO matches (eventID) VALUES (?)", eID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func addAlliances(count int, matchID int, isBlue bool) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		_, err := s.DB.Exec("INSERT INTO alliances(matchID, score, isBlue, team1, team2, team3) VALUES (?, ?, ?, ?, ?, ?)", matchID, ((i + 1) * 25), isBlue, 0, 0, 0)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func addReports(count int, allianceID int, team int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		_, err := s.DB.Exec("INSERT INTO reports(allianceID, teamNumber, score) VALUES (?, ?, ?)", allianceID, (allianceID * 25), team)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		_, err = s.DB.Exec("UPDATE alliances SET team1=? WHERE id=?", team, allianceID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func TestEmptyEventTable(t *testing.T) {
	setupTables()

	req, _ := http.NewRequest("GET", "/events", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetEventMultiple(t *testing.T) {
	setupTables()

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
	setupTables()

	addEvents(1)
	addMatches(5, 1)

	req, _ := http.NewRequest("GET", "/events/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	fe := fullEvent{}

	json.Unmarshal(response.Body.Bytes(), &fe)

	if fe.Matches[2].EventID != 1 {
		t.Errorf("Expected event ID to be '1'. Got '%v' instead.", fe.Matches[0].EventID)
	}
}

func TestAddValidReport(t *testing.T) {
	setupTables()
	addEvents(1)
	addMatches(1, 1)

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	req, _ := http.NewRequest("POST", "/events/1/1", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)
	report := reportData{}
	json.Unmarshal(response.Body.Bytes(), &report)

	if report.Alliance != "red" {
		t.Errorf("Expected report alliance to be 'red'. Got '%v' instead.", report.Alliance)
	}
	if report.Team != 2733 {
		t.Errorf("Expected report team number to be '2733'. Got '%v' instead.", report.Team)
	}

	auto := &autoReport{CrossedLine: true, DeliveredGear: true, Fuel: 10}
	teleop := &teleopReport{Climbed: true, Gears: 3, Fuel: 10}

	if report.Auto != *auto {
		t.Errorf("Auto section of report was not as expected")
	}
	if report.Teleop != *teleop {
		t.Errorf("Teleop section of report was not as expected")
	}
}

func TestAddInvalidReport(t *testing.T) {
	setupTables()
	addEvents(1)
	addMatches(1, 1)

	payload := []byte(`{ "allince": 12, "tea": 2733,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": 9, "gears": 3, "fuel": 10 }}`)

	req, _ := http.NewRequest("POST", "/events/1/1", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestAddExistingReport(t *testing.T) {
	setupTables()
	addEvents(1)
	addMatches(1, 1)

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	req, _ := http.NewRequest("POST", "/events/1/1", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("POST", "/events/1/1", bytes.NewBuffer(payload))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusConflict, response.Code)
}

func TestUpdateReport(t *testing.T) {
	setupTables()
	addEvents(1)
	addMatches(1, 1)
	addAlliances(1, 1, true)
	addReports(1, 1, 2733)

	payload := []byte(`{ "alliance": "blue", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	req, _ := http.NewRequest("PUT", "/events/1/1/2733", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	report := reportData{}
	json.Unmarshal(response.Body.Bytes(), &report)

	if report.Alliance != "blue" {
		t.Errorf("Expected updated report alliance to be 'blue'. Got '%v' instead.", report.Alliance)
	}
	if report.Team != 2733 {
		t.Errorf("Expected updated report team number to be '2733'. Got '%v' instead.", report.Team)
	}

	auto := &autoReport{CrossedLine: true, DeliveredGear: true, Fuel: 10}
	teleop := &teleopReport{Climbed: true, Gears: 3, Fuel: 10}

	if report.Auto != *auto {
		t.Errorf("Auto section of report was not as expected")
	}
	if report.Teleop != *teleop {
		t.Errorf("Teleop section of report was not as expected")
	}
}

func TestUpdateNonexistentReport(t *testing.T) {
	setupTables()

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	req, _ := http.NewRequest("PUT", "/events/1/1/2733", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}
