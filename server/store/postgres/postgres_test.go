package sqlite3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Pigmice2733/scouting-backend/server"
	"github.com/Pigmice2733/scouting-backend/server/store"
)

var s *server.Server
var rawStore *service

func TestMain(m *testing.M) {
	// testing connects to a postgres db created by docker, check readme for more information

	store, err := NewFromOptions(Options{User: os.Getenv("POSTGRES_1_ENV_POSTGRES_USER"), Pass: os.Getenv("POSTGRES_1_ENV_POSTGRES_PASSWORD"), Host: os.Getenv("POSTGRES_1_PORT_5432_TCP_ADDR"), Port: 5432, DBName: os.Getenv("POSTGRES_1_ENV_POSTGRES_DB"), SSLMode: "disable"})
	if err != nil {
		log.Fatalf("error creating database: %v\n", err)
	}

	var ok bool
	if rawStore, ok = store.(*service); !ok {
		log.Fatalf("cannot cast store to private type for testing")
	}

	s = server.New(store, os.Stdout, "", "dev")
	exitCode := m.Run()
	os.Exit(exitCode)
}

func setupTables() error {
	if err := ensureTableExists(eventTableCreationQuery); err != nil {
		return fmt.Errorf("error: ensuring event table exists: %v", err)
	}
	if err := ensureTableExists(matchTableCreationQuery); err != nil {
		return fmt.Errorf("error: ensuring match table exists: %v", err)
	}
	if err := ensureTableExists(allianceTableCreationQuery); err != nil {
		return fmt.Errorf("error: ensuring alliance table exists: %v", err)
	}
	if err := ensureTableExists(reportTableCreationQuery); err != nil {
		return fmt.Errorf("error: ensuring report table exists: %v", err)
	}

	// Order matters!
	if err := rawStore.clearTable("reports"); err != nil {
		return fmt.Errorf("error: creating table: %v", err)
	}
	if err := rawStore.clearTable("alliances"); err != nil {
		return fmt.Errorf("error: creating table: %v", err)
	}
	if err := rawStore.clearTable("matches"); err != nil {
		return fmt.Errorf("error: creating table: %v", err)
	}
	if err := rawStore.clearTable("events"); err != nil {
		return fmt.Errorf("error: creating table: %v", err)
	}

	return nil
}

func ensureTableExists(creationQuery string) error {
	_, err := rawStore.db.Exec(creationQuery)
	return err
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("error: expected response code %d, got %d\n", expected, actual)
	}
}

func addEvent() (string, error) {
	_, err := rawStore.db.Exec("INSERT INTO events (key, name, date) VALUES ($1, $2, $3)", "Evnt1", "Event 1", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return "", err
	}
	return "Evnt1", nil
}

func addMatch(eKey string) (string, error) {
	_, err := rawStore.db.Exec("INSERT INTO matches (key, eventKey) VALUES ($1, $2)", "Mtch1", eKey)
	if err != nil {
		return "", err
	}
	return "Mtch1", nil
}

func addFullMatch(eKey string, winningAlliance string) (string, error) {
	_, err := rawStore.db.Exec("INSERT INTO matches (key, eventKey, winningAlliance) VALUES ($1, $2, $3)", "FullMatch", eKey, winningAlliance)
	if err != nil {
		return "", err
	}
	return "FullMatch", nil
}

func addAlliance(matchKey string, isBlue bool) (id int, err error) {
	err = rawStore.db.QueryRow("INSERT INTO alliances(matchKey, score, isBlue, team1, team2, team3) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", matchKey, 451, isBlue, 0, 0, 0).Scan(&id)
	return id, nil
}

func addReports(allianceID int, team int) error {
	_, err := rawStore.db.Exec("INSERT INTO reports(allianceID, teamNumber, score) VALUES ($1, $2, $3)", allianceID, team, (allianceID * 25))
	if err != nil {
		return err
	}

	_, err = rawStore.db.Exec("UPDATE alliances SET team1=$1 WHERE id=$2", team, allianceID)
	if err != nil {
		return err
	}
	return nil
}

func TestEmptyEventTable(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

	req, _ := http.NewRequest("GET", "/events", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("error: expected an empty array. Got %v", body)
	}
}

func TestGetEventMultiple(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

	_, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}

	req, _ := http.NewRequest("GET", "/events", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	d := []store.Event{}

	if err := json.Unmarshal(response.Body.Bytes(), &d); err != nil {
		t.Errorf("error: unmarshalling body: %v\n", err)
	}

	if d[0].Name != "Event 1" {
		t.Errorf("error: expected event name to be 'Event 1'. Got '%v' instead.", d[0].Name)
	}
}

func TestGetMatchMultiple(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = addMatch(eKey)
	if err != nil {
		t.Errorf(err.Error())
	}

	endpoint := fmt.Sprintf("/events/%s", eKey)

	req, _ := http.NewRequest("GET", endpoint, nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	fe := store.FullEvent{}

	if err := json.NewDecoder(response.Body).Decode(&fe); err != nil {
		t.Errorf("error: unmarshalling body: %v\n", err)
	}

	if fe.Matches[0].EventKey != eKey {
		t.Errorf("error: expected event key to be '%v'. Got '%v' instead.", eKey, fe.Matches[0].EventKey)
	}
}

func TestGetMatchData(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

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

	endpoint := fmt.Sprintf("/events/%s/%s", eKey, mKey)

	req, _ := http.NewRequest("GET", endpoint, nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	fm := store.FullMatch{}

	if err := json.NewDecoder(response.Body).Decode(&fm); err != nil {
		t.Errorf("error: unmarshalling body: %v\n", err)
	}

	if fm.EventKey != eKey {
		t.Errorf("error: expected event key to be '%v'. Got '%v' instead.", eKey, fm.EventKey)
	}
	if fm.WinningAlliance != "blue" {
		t.Errorf("error: expected match winning alliance to be 'blue'. Got '%v' instead.", fm.WinningAlliance)
	}
}

func TestAddValidReport(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

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

	endpoint := fmt.Sprintf("/events/%s/%s", eKey, mKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)
	report := store.ReportData{}

	if err := json.NewDecoder(response.Body).Decode(&report); err != nil {
		t.Errorf("error: unmarshalling body: %v\n", err)
	}

	if report.Alliance != "red" {
		t.Errorf("error: expected report alliance to be 'red'. Got '%v' instead.", report.Alliance)
	}
	if report.Team != 2733 {
		t.Errorf("error: expected report team number to be '2733'. Got '%v' instead.", report.Team)
	}

	auto := &store.AutoReport{CrossedLine: true, DeliveredGear: true, Fuel: 10}
	teleop := &store.TeleopReport{Climbed: true, Gears: 3, Fuel: 10}

	if report.Auto != *auto {
		t.Errorf("error: auto section of report was not as expected")
	}
	if report.Teleop != *teleop {
		t.Errorf("error: teleop section of report was not as expected")
	}
}

func TestAddInvalidReport(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

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

	endpoint := fmt.Sprintf("/events/%v/%v", eKey, mKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestPostReportFakeMatch(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

	eKey, err := addEvent()
	if err != nil {
		t.Errorf(err.Error())
	}

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	endpoint := fmt.Sprintf("/event/%v/1", eKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestPostExistingReport(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

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

	endpoint := fmt.Sprintf("/events/%v/%v", eKey, mKey)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusConflict, response.Code)
}

func TestUpdateReport(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

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

	endpoint := fmt.Sprintf("/events/%v/%v/2733", eKey, mKey)

	req, _ := http.NewRequest("PUT", endpoint, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	report := store.ReportData{}

	if err := json.NewDecoder(response.Body).Decode(&report); err != nil {
		t.Errorf("error: unmarshalling body: %v\n", err)
	}

	if report.Alliance != "blue" {
		t.Errorf("error: expected updated report alliance to be 'blue'. Got '%v' instead.", report.Alliance)
	}
	if report.Team != 2733 {
		t.Errorf("error: expected updated report team number to be '2733'. Got '%v' instead.", report.Team)
	}

	auto := &store.AutoReport{CrossedLine: true, DeliveredGear: true, Fuel: 10}
	teleop := &store.TeleopReport{Climbed: true, Gears: 3, Fuel: 10}

	if report.Auto != *auto {
		t.Errorf("error: auto section of report was not as expected")
	}
	if report.Teleop != *teleop {
		t.Errorf("error: teleop section of report was not as expected")
	}
}

func TestUpdateNonexistentReport(t *testing.T) {
	if err := setupTables(); err != nil {
		t.Fatalf("error: setting up tables: %v\n", err)
	}

	payload := []byte(`{ "alliance": "red", "team": 2733, "score": 451,
		"auto": { "crossedLine": true, "deliveredGear": true, "fuel": 10 },
		"teleop": { "climbed": true, "gears": 3, "fuel": 10 }}`)

	req, _ := http.NewRequest("PUT", "/events/Evnt1/1/2733", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}
