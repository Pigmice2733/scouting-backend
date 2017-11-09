// data_models_test.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var s *Server

func TestMain(m *testing.M) {
	s = New("file::memory:?mode=memory&cache=shared", os.Stdout, "", "dev")
	exitCode := m.Run()
	os.Exit(exitCode)
}

func setupTables() {
	ensureTableExists(eventTableCreationQuery)
	ensureTableExists(matchTableCreationQuery)
	ensureTableExists(allianceTableCreationQuery)
	ensureTableExists(reportTableCreationQuery)
	// Order matters!
	clearTable("reports")
	clearTable("alliances")
	clearTable("matches")
	clearTable("events")
}

func ensureTableExists(creationQuery string) {
	s.DB.Exec(creationQuery)
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

func addEvent() (string, error) {
	_, err := s.DB.Exec("INSERT INTO events (key, name, date) VALUES (?, ?, ?)", "Evnt1", "Event 1", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return "", err
	}
	return "Evnt1", nil
}

func addMatch(eKey string) (string, error) {
	_, err := s.DB.Exec("INSERT INTO matches (key, eventKey, number) VALUES (?, ?, ?)", "Mtch1", eKey, -1)
	if err != nil {
		return "", err
	}
	return "Mtch1", nil
}

func addFullMatch(eKey string, winningAlliance string) (string, error) {
	_, err := s.DB.Exec("INSERT INTO matches (key, eventKey, number, winningAlliance) VALUES (?, ?, ?, ?)", "FullMatch", eKey, -1, winningAlliance)
	if err != nil {
		return "", err
	}
	return "FullMatch", nil
}

func addAlliance(matchKey string, isBlue bool) (int, error) {
	res, err := s.DB.Exec("INSERT INTO alliances(matchKey, score, isBlue, team1, team2, team3) VALUES (?, ?, ?, ?, ?, ?)", matchKey, 451, isBlue, 0, 0, 0)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func addReports(allianceID int, team int) error {
	_, err := s.DB.Exec("INSERT INTO reports(allianceID, teamNumber, score) VALUES (?, ?, ?)", allianceID, team, (allianceID * 25))
	if err != nil {
		return err
	}

	_, err = s.DB.Exec("UPDATE alliances SET team1=? WHERE id=?", team, allianceID)
	if err != nil {
		return err
	}
	return nil
}

func TestEmptyEventTable(t *testing.T) {
	setupTables()

	req, _ := http.NewRequest("GET", "/events/", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %v", body)
	}
}

func TestGetEventMultiple(t *testing.T) {
	setupTables()

	_, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}

	req, _ := http.NewRequest("GET", "/events/", nil)
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

	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = addMatch(eKey)
	if err != nil {
		t.Errorf(err.Error())
	}

	endpoint := fmt.Sprintf("/events/%s/", eKey)

	req, _ := http.NewRequest("GET", endpoint, nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	fe := fullEvent{}

	json.Unmarshal(response.Body.Bytes(), &fe)

	if fe.Matches[0].EventKey != eKey {
		t.Errorf("Expected event key to be '%v'. Got '%v' instead.", eKey, fe.Matches[0].EventKey)
	}
}

func TestGetMatchData(t *testing.T) {
	setupTables()
	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}
	mKey, err := addFullMatch(eKey, "blue")
	if err != nil {
		t.Errorf(err.Error())
	}
	id, err := addAlliance(mKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = addReports(id, 2733)
	if err != nil {
		t.Errorf(err.Error())
	}

	endpoint := fmt.Sprintf("/events/%s/%s/", eKey, mKey)

	req, _ := http.NewRequest("GET", endpoint, nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	fm := fullMatch{}

	json.Unmarshal(response.Body.Bytes(), &fm)

	if fm.EventKey != eKey {
		t.Errorf("Expected event key to be '%v'. Got '%v' instead.", eKey, fm.EventKey)
	}
	if fm.WinningAlliance != "blue" {
		t.Errorf("Expected match winning alliance to be 'blue'. Got '%v' instead.", fm.WinningAlliance)
	}
}

func TestAddValidReport(t *testing.T) {
	setupTables()
	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}
	mKey, err := addMatch(eKey)
	if err != nil {
		t.Errorf(err.Error())
	}

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	endpoint := fmt.Sprintf("/events/%s/%s/", eKey, mKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
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
	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}
	mKey, err := addMatch(eKey)
	if err != nil {
		t.Errorf(err.Error())
	}

	payload := []byte(`{ "allince": 12, "tea": 2733,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": 9, "gears": 3, "fuel": 10 }}`)

	endpoint := fmt.Sprintf("/events/%v/%v/", eKey, mKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestPostReportFakeMatch(t *testing.T) {
	setupTables()
	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	endpoint := fmt.Sprintf("/event/%v/1/", eKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestPostExistingReport(t *testing.T) {
	setupTables()
	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}
	mKey, err := addMatch(eKey)
	if err != nil {
		t.Errorf(err.Error())
	}

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	endpoint := fmt.Sprintf("/events/%v/%v/", eKey, mKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusConflict, response.Code)
}

func TestUpdateReport(t *testing.T) {
	setupTables()
	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}
	mKey, err := addMatch(eKey)
	if err != nil {
		t.Errorf(err.Error())
	}
	id, err := addAlliance(mKey, true)
	if err != nil {
		t.Errorf(err.Error())
	}
	err = addReports(id, 2733)
	if err != nil {
		t.Errorf(err.Error())
	}

	payload := []byte(`{ "alliance": "blue", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	endpoint := fmt.Sprintf("/events/%v/%v/2733/", eKey, mKey)

	req, _ := http.NewRequest("PUT", endpoint, bytes.NewBuffer(payload))
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

	req, _ := http.NewRequest("PUT", "/events/Evnt1/1/2733/", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}
