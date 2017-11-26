package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Pigmice2733/scouting-backend/server/store"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var globalStore store.Service
var rawDB *sql.DB

var options = Options{User: os.Getenv("POSTGRES_1_ENV_POSTGRES_USER"), Pass: os.Getenv("POSTGRES_1_ENV_POSTGRES_PASSWORD"), Host: os.Getenv("POSTGRES_1_PORT_5432_TCP_ADDR"), Port: 5432, DBName: os.Getenv("POSTGRES_1_ENV_POSTGRES_DB"), SSLMode: "disable", MaxConnections: 60, StatementTimeout: 5000}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	connectionsClosed := teardown()
	if !connectionsClosed {
		os.Exit(1)
	}
	os.Exit(code)
}

func setup() {
	var err error
	globalStore, err = NewFromOptions(options)
	if err != nil {
		fmt.Printf("error connecting to database: %v\n", err)
		os.Exit(1)
	}

	if s, ok := globalStore.(*service); ok {
		rawDB = s.db
		rawDB.SetMaxOpenConns(20)
		rawDB.SetMaxIdleConns(0)
	} else {
		fmt.Printf("error getting underlying service: %v\n", err)
	}
}

func teardown() bool {
	closed := checkConnections()

	err := globalStore.Close()
	if err != nil {
		log.Panicf("error closing db connection %v", err)
	}
	rawDB = nil
	return closed
}

func checkConnections() bool {
	time.Sleep(time.Millisecond)
	if rawDB == nil {
		// DB connection failed
		return true
	}
	open := rawDB.Stats().OpenConnections
	if open > 0 {
		fmt.Printf("tests failed to close %d connections\n", open)
		return false
	}
	return true
}

var eventCaseKeyInc = &keyInc{}
var eventCases = []store.Event{
	store.Event{Key: eventCaseKeyInc.Next(), Name: "", Date: time.Unix(10000, 0)},
	store.Event{Key: eventCaseKeyInc.Next(), Name: "my-name-is-who", Date: time.Unix(10000, 0)},
}

var eventKey = eventCaseKeyInc.Next()

var matchCaseKeyInc = &keyInc{}
var matchCases = []struct {
	m         store.Match
	insertErr bool
	getErr    bool
}{
	{store.Match{Key: matchCaseKeyInc.Next(), EventKey: eventKey, PredictedTime: time.Unix(45678, 0).UTC(), ActualTime: time.Unix(43786, 0).UTC(), WinningAlliance: "red"}, false, false},
	{store.Match{Key: matchCaseKeyInc.Next(), EventKey: eventKey, PredictedTime: time.Unix(36573, 0).UTC(), ActualTime: time.Unix(43786, 0).UTC(), WinningAlliance: "blue"}, false, false},
	{store.Match{Key: matchCaseKeyInc.Next(), EventKey: "i-dont-exist", PredictedTime: time.Unix(36573, 0).UTC(), ActualTime: time.Unix(43786, 0).UTC(), WinningAlliance: "red"}, true, true},
}

var matchKey = matchCaseKeyInc.Next()

var allianceCases = []struct {
	a         store.Alliance
	insertErr bool
	getErr    bool
}{
	{store.Alliance{MatchKey: matchKey, IsBlue: false, Score: 100}, false, false},
	{store.Alliance{MatchKey: matchKey, IsBlue: true, Score: 200}, false, false},
	{store.Alliance{MatchKey: "i-dont-exist"}, true, true},
}

var teamCases = []struct {
	test      store.TeamInAlliance
	insertErr bool
	getErr    bool
}{
	{store.TeamInAlliance{Number: "2733", PredictedContribution: "some", ActualContribution: "nothing"}, false, false},
	{store.TeamInAlliance{Number: "1418", ActualContribution: ""}, false, false},
	{store.TeamInAlliance{Number: "1540b", PredictedContribution: "a lot"}, false, false},
}

var reportCases = []struct {
	rd        store.ReportData
	insertErr bool
	getErr    bool
}{
	{store.ReportData{
		Reporter: "frank", Alliance: "red", Team: "2733", Score: 451, Auto: store.AutoReport{
			CrossedLine: true, DeliveredGear: true, Fuel: 10,
		}, Teleop: store.TeleopReport{
			Climbed: true, Gears: 3, Fuel: 10,
		},
	}, false, false},
	{store.ReportData{
		Reporter: "brendan", Alliance: "blue", Team: "2471", Score: 200, Auto: store.AutoReport{
			CrossedLine: false, DeliveredGear: false, Fuel: 5,
		}, Teleop: store.TeleopReport{
			Climbed: false, Gears: 1, Fuel: 2,
		},
	}, false, false},
}

const exampleHash = "$2b$12$NtJjdgSOJIdwDOWvRgIX7.w7PK2GMLT4OdxuYVDzYxHIAbtX5ROPK"

var userCases = []struct {
	user      store.User
	insertErr bool
	getErr    bool
}{
	{user: store.User{Username: "frank", HashedPassword: exampleHash}, insertErr: false, getErr: false},
	{user: store.User{Username: "brendan", HashedPassword: exampleHash}, insertErr: false, getErr: false},
	{user: store.User{Username: "caleb", HashedPassword: exampleHash}, insertErr: false, getErr: false},
	{user: store.User{Username: "frank", HashedPassword: exampleHash}, insertErr: true, getErr: true},
}

func clearTables(tables ...string) error {
	for _, table := range tables {
		_, err := rawDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return err
		}
	}
	return nil
}

func roughEventEquality(e1, e2 store.Event) bool {
	return (e1.Name == e2.Name) && (e1.Key == e2.Key) && (e1.Date.Unix() == e2.Date.Unix())
	// You have to convert to unix time because otherwise you get floating point precision stuff
	// that makes them "not equal" even though they practically are, and we are only testing for
	// down to the second, any precision beyond that is unimportant. This is also why we have to
	// test all the fields seperately.
}

// keyInc provides an easy way to generate unique keys sequentially for testing
type keyInc struct {
	N int
}

func (ki *keyInc) Next() string {
	ki.N++
	return ki.Key()
}

func (ki *keyInc) Key() string {
	return fmt.Sprintf("%x", ki.N)
}

func TestGetEventAndGetEvents(t *testing.T) {
	err := clearTables("reports", "teamsInAlliance", "alliances", "matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	for i, c := range eventCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err := rawDB.Exec("INSERT INTO events VALUES($1, $2, $3)", c.Key, c.Name, c.Date)
		assert.Equal(t, nil, err, caseString)

		e, err := globalStore.GetEvent(c.Key)
		assert.Equal(t, nil, err, caseString)

		equality := roughEventEquality(c, e)
		assert.Equal(t, true, equality, caseString)
	}

	events, err := globalStore.GetEvents()
	assert.Equal(t, nil, err)

	for _, eCase := range eventCases {
		var exists bool
		for _, eGot := range events {
			if roughEventEquality(eGot, eCase) {
				exists = true
			}
		}

		assert.Equal(t, true, exists)
	}
	// shhh O(n^2) is fine for testing
}

func TestCreateEventAndUpdateEvents(t *testing.T) {
	_, err := rawDB.Exec("DELETE FROM events")
	assert.Equal(t, nil, err, "clearing events table")

	for i, c := range eventCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		err := globalStore.CreateEvent(c)
		assert.Equal(t, nil, err, caseString)

		var e store.Event
		err = rawDB.QueryRow("SELECT key, name, date FROM events WHERE key = $1", c.Key).Scan(
			&e.Key, &e.Name, &e.Date)
		assert.Equal(t, nil, err, caseString)

		equality := roughEventEquality(e, c)
		assert.Equal(t, true, equality, caseString)
	}

	var updatedEvents []store.Event
	for _, event := range eventCases {
		updatedEvents = append(updatedEvents, store.Event{
			Key: event.Key, Name: event.Name + "a", Date: event.Date.Add(time.Minute)})
	}
	// for testing the SERT part of UPSERT
	updatedEvents = append(updatedEvents, store.Event{
		Key: eventCaseKeyInc.Next(), Name: "asdf", Date: time.Unix(1000, 0)})

	errs := globalStore.UpdateEvents(updatedEvents)
	assert.Equal(t, 0, len(errs))

	for _, updatedEvent := range updatedEvents {
		var dbEvent store.Event
		err := rawDB.QueryRow("SELECT key, name, date FROM events WHERE key = $1", updatedEvent.Key).Scan(
			&dbEvent.Key, &dbEvent.Name, &dbEvent.Date)
		assert.Equal(t, nil, err)

		equality := roughEventEquality(updatedEvent, dbEvent)
		assert.Equal(t, true, equality)
	}
}

func TestGetMatchAndGetMatches(t *testing.T) {
	err := clearTables("matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	_, err = rawDB.Exec("INSERT INTO events VALUES ($1, $2, $3)", eventKey, "", time.Unix(10000, 0))
	assert.Equal(t, nil, err, "inserting event for testing matches")

	for i, c := range matchCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err := rawDB.Exec("INSERT INTO matches (key, eventKey, predictedTime, actualTime, winningAlliance) VALUES ($1, $2, $3, $4, $5)",
			c.m.Key, c.m.EventKey, c.m.PredictedTime, c.m.ActualTime, c.m.WinningAlliance)
		assert.Equal(t, c.insertErr, err != nil, caseString)
		var allianceID int
		// Blue alliance and team
		err = rawDB.QueryRow("INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id",
			c.m.Key, true, 100).Scan(&allianceID)
		assert.Equal(t, c.insertErr, err != nil, caseString)
		_, err = rawDB.Exec("INSERT INTO teamsInAlliance (number, allianceID, predictedContribution, actualContribution) VALUES($1, $2, $3, $4)", "1540b", allianceID, "predicted", "actual")
		assert.Equal(t, c.insertErr, err != nil, caseString)
		// Red alliance and team
		err = rawDB.QueryRow("INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id",
			c.m.Key, false, 100).Scan(&allianceID)
		assert.Equal(t, c.insertErr, err != nil, caseString)
		_, err = rawDB.Exec("INSERT INTO teamsInAlliance (number, allianceID, predictedContribution, actualContribution) VALUES($1, $2, $3, $4)", "frc2733", allianceID, "failure", "PIGMICE!")
		assert.Equal(t, c.insertErr, err != nil, caseString)

		m, err := globalStore.GetMatch(c.m.EventKey, c.m.Key)
		assert.Equal(t, c.getErr, err != nil, caseString)

		if !c.insertErr && !c.getErr {
			assert.Equal(t, c.m, m, caseString)
		}
	}

	matches, err := globalStore.GetAllMatchData(eventKey)
	assert.Equal(t, nil, err)

	for _, mCase := range matchCases {
		if mCase.insertErr || mCase.getErr || mCase.m.EventKey != eventKey {
			continue
		}

		mCase.m.BlueAlliance = store.Alliance{MatchKey: mCase.m.Key, IsBlue: true, Score: 100}
		mCase.m.RedAlliance = store.Alliance{MatchKey: mCase.m.Key, IsBlue: false, Score: 100}
		mCase.m.BlueAlliance.Teams = []store.TeamInAlliance{store.TeamInAlliance{Number: "1540b", PredictedContribution: "predicted", ActualContribution: "actual"}}
		mCase.m.RedAlliance.Teams = []store.TeamInAlliance{store.TeamInAlliance{Number: "frc2733", PredictedContribution: "failure", ActualContribution: "PIGMICE!"}}

		var exists bool
		for _, mGot := range matches {
			mCase.m.BlueAlliance.ID = mGot.BlueAlliance.ID
			mCase.m.BlueAlliance.Teams[0].AllianceID = mGot.BlueAlliance.ID
			mCase.m.RedAlliance.ID = mGot.RedAlliance.ID
			mCase.m.RedAlliance.Teams[0].AllianceID = mGot.RedAlliance.ID
			if reflect.DeepEqual(mGot, mCase.m) {
				exists = true
			}
		}
		assert.Equal(t, true, exists)
	}
	// shhh O(n^2) is fine for testing
}

func TestCheckMatchExistence(t *testing.T) {
	err := clearTables("teamsInAlliance", "alliances", "matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	_, err = rawDB.Exec("INSERT INTO events (key, name, date) VALUES ($1, $2, $3)", eventKey, "", time.Unix(10000, 0))
	assert.Equal(t, nil, err, "inserting event for testing matches")

	_, err = rawDB.Exec("INSERT INTO matches (key, eventKey, winningAlliance) VALUES ($1, $2, $3)", matchKey, eventKey, "red")
	assert.Equal(t, nil, err, "inserting match for testing match existence check")

	exists, err := globalStore.CheckMatchExistence(eventKey, matchKey)
	assert.Equal(t, true, exists, "CheckMatchExistence failed to find inserted match")

	exists, err = globalStore.CheckMatchExistence(eventKey, "frc_fakeMatchKey")
	assert.Equal(t, false, exists, "CheckMatchExistence found non-existent match")
}

func TestCreateMatchAndUpdateMatches(t *testing.T) {
	err := clearTables("matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	_, err = rawDB.Exec("INSERT INTO events VALUES ($1, $2, $3)", eventKey, "", time.Unix(10000, 0))
	assert.Equal(t, nil, err, "inserting event for testing matches")

	for i, c := range matchCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		err := globalStore.CreateMatch(c.m)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		if c.insertErr {
			continue
		}

		var m store.Match
		err = rawDB.QueryRow("SELECT key, eventKey, predictedTime, actualTime, winningAlliance FROM matches WHERE key = $1 AND eventKey = $2",
			c.m.Key, c.m.EventKey).Scan(&m.Key, &m.EventKey, &m.PredictedTime, &m.ActualTime, &m.WinningAlliance)

		assert.Equal(t, c.m, m, caseString)
	}

	var updatedMatches []store.Match
	for _, match := range matchCases {
		updatedMatches = append(updatedMatches, store.Match{
			Key: match.m.Key, EventKey: eventKey, PredictedTime: match.m.PredictedTime.Add(time.Hour), WinningAlliance: "red",
			BlueAlliance: store.Alliance{MatchKey: match.m.Key, IsBlue: true, Score: 2733,
				Teams: []store.TeamInAlliance{store.TeamInAlliance{Number: "frc2733b"}}},
			RedAlliance: store.Alliance{MatchKey: match.m.Key, IsBlue: false, Score: 0000},
		})
	}
	// for testing the SERT part of UPSERT
	updatedMatches = append(updatedMatches, store.Match{
		Key: "frc2733TheBest", EventKey: eventKey, WinningAlliance: "green",
		BlueAlliance: store.Alliance{MatchKey: "frc2733TheBest", IsBlue: true, Score: 656},
		RedAlliance:  store.Alliance{MatchKey: "frc2733TheBest", IsBlue: false, Score: 421}})

	errs := globalStore.UpdateMatches(updatedMatches)
	assert.Equal(t, 0, len(errs))

	for _, updatedMatch := range updatedMatches {
		var dbMatch store.Match
		dbMatch, err := retrieveCompleteMatch(updatedMatch.EventKey, updatedMatch.Key)
		assert.Equal(t, nil, err)

		equality := reflect.DeepEqual(updatedMatch, dbMatch)
		assert.Equal(t, true, equality)
	}
}

func TestGetAlliance(t *testing.T) {
	err := clearTables("teamsInAlliance", "alliances", "matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	_, err = rawDB.Exec("INSERT INTO events (key, name, date) VALUES ($1, $2, $3)", eventKey, "", time.Unix(10000, 0))
	assert.Equal(t, nil, err, "inserting event for testing matches")

	_, err = rawDB.Exec("INSERT INTO matches (key, eventKey, winningAlliance) VALUES ($1, $2, $3)", matchKey, eventKey, "red")
	assert.Equal(t, nil, err, "inserting match for testing alliances")

	for i, c := range allianceCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err := rawDB.Exec(
			"INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3)",
			c.a.MatchKey, c.a.IsBlue, c.a.Score)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		if c.insertErr {
			continue
		}

		a, _, err := globalStore.GetAlliance(c.a.MatchKey, c.a.IsBlue)
		a.ID = 0
		assert.Equal(t, c.getErr, err != nil, caseString)

		if !c.insertErr && !c.getErr {
			assert.Equal(t, c.a, a, caseString)
		}
	}
}

func TestCreateAllianceAndUpdateAlliance(t *testing.T) {
	err := clearTables("teamsInAlliance", "alliances", "matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	_, err = rawDB.Exec("INSERT INTO events VALUES ($1, $2, $3)", eventKey, "", time.Unix(10000, 0))
	assert.Equal(t, nil, err, "inserting event for testing matches")

	_, err = rawDB.Exec("INSERT INTO matches (key, eventKey, winningAlliance) VALUES ($1, $2, $3)", matchKey, eventKey, "red")
	assert.Equal(t, nil, err, "inserting match for testing alliances")

	for i, c := range allianceCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err := globalStore.CreateAlliance(c.a)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		if c.insertErr {
			continue
		}

		var a store.Alliance
		err = rawDB.QueryRow(
			"SELECT matchKey, isBlue, score FROM alliances WHERE matchKey = $1 AND isBlue = $2",
			c.a.MatchKey, c.a.IsBlue).Scan(&a.MatchKey, &a.IsBlue, &a.Score)
		assert.Equal(t, c.getErr, err != nil, caseString)

		assert.Equal(t, c.a, a, caseString)

		a = store.Alliance{MatchKey: a.MatchKey, IsBlue: a.IsBlue, Score: a.Score * 2}
		err = globalStore.UpdateAlliance(a)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		var updated store.Alliance
		err = rawDB.QueryRow(
			"SELECT matchKey, isBlue, score FROM alliances WHERE matchKey = $1 AND isBlue = $2",
			a.MatchKey, a.IsBlue).Scan(&updated.MatchKey, &updated.IsBlue, &updated.Score)
		assert.Equal(t, c.getErr, err != nil, caseString)

		assert.Equal(t, a, updated, caseString)
	}
}

func TestCreateReportAndUpdateReport(t *testing.T) {
	err := clearTables("reports", "teamsInAlliance", "alliances", "matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	var eventKeyInc keyInc
	var matchKeyInc keyInc

	for i, c := range reportCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err = rawDB.Exec("INSERT INTO events VALUES ($1, $2, $3)", eventKeyInc.Next(), "", time.Unix(10000, 0))
		assert.Equal(t, nil, err, "inserting event for testing matches")

		_, err = rawDB.Exec("INSERT INTO matches(key, eventKey, winningAlliance) VALUES ($1, $2, $3)", matchKeyInc.Next(), eventKeyInc.Key(), c.rd.Alliance)
		assert.Equal(t, nil, err, "inserting match for testing alliances")

		var allianceID int
		err = rawDB.QueryRow(
			"INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id",
			matchKeyInc.Key(), c.rd.Alliance == "blue", 100).Scan(&allianceID)

		err := globalStore.CreateReport(c.rd, allianceID)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		if c.insertErr {
			continue
		}

		var rd store.ReportData
		err = rawDB.QueryRow(
			"SELECT reporter, teamNumber, score, crossedLine, deliveredGear, autoFuel, climbed, gears, teleopFuel FROM reports WHERE allianceID = $1",
			allianceID).Scan(&rd.Reporter, &rd.Team, &rd.Score, &rd.Auto.CrossedLine, &rd.Auto.DeliveredGear, &rd.Auto.Fuel,
			&rd.Teleop.Climbed, &rd.Teleop.Gears, &rd.Teleop.Fuel)
		assert.Equal(t, c.getErr, err != nil, caseString)

		rd.Alliance = c.rd.Alliance

		assert.Equal(t, c.rd, rd, caseString)

		rd = store.ReportData{
			Alliance: rd.Alliance, Team: rd.Team, Score: rd.Score * 2,
			Auto:   store.AutoReport{CrossedLine: !rd.Auto.CrossedLine, DeliveredGear: !rd.Auto.DeliveredGear, Fuel: rd.Auto.Fuel * 2},
			Teleop: store.TeleopReport{Climbed: !rd.Teleop.Climbed, Gears: rd.Teleop.Gears * 2, Fuel: rd.Teleop.Fuel * 2},
		}
		err = globalStore.UpdateReport(rd, allianceID)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		var updated store.ReportData
		err = rawDB.QueryRow(
			"SELECT reporter, teamNumber, score, crossedLine, deliveredGear, autoFuel, climbed, gears, teleopFuel FROM reports WHERE allianceID = $1",
			allianceID).Scan(&updated.Reporter, &updated.Team, &updated.Score, &updated.Auto.CrossedLine, &updated.Auto.DeliveredGear, &updated.Auto.Fuel,
			&updated.Teleop.Climbed, &updated.Teleop.Gears, &updated.Teleop.Fuel)
		assert.Equal(t, c.getErr, err != nil, caseString)

		updated.Alliance = rd.Alliance

		assert.Equal(t, rd, updated, caseString)
	}

}

func TestGetUserAndGetUsers(t *testing.T) {
	_, err := rawDB.Exec("DELETE FROM users")
	assert.Equal(t, nil, err, "clearing users table")

	for i, c := range userCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err := rawDB.Exec("INSERT INTO users VALUES($1, $2)", c.user.Username, c.user.HashedPassword)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		if c.insertErr {
			continue
		}

		user, err := globalStore.GetUser(c.user.Username)
		assert.Equal(t, c.getErr, err != nil, caseString)

		assert.Equal(t, c.user, user, caseString)
	}

	users, err := globalStore.GetUsers()
	assert.Equal(t, nil, err)

	for _, uCase := range userCases {
		if uCase.insertErr || uCase.getErr {
			continue
		}

		var exists bool
		for _, uGot := range users {
			if uGot == uCase.user {
				exists = true
			}
		}

		assert.Equal(t, true, exists)
	}
	// shhh O(n^2) is fine for testing
}

func TestCreateUser(t *testing.T) {
	_, err := rawDB.Exec("DELETE FROM users")
	assert.Equal(t, nil, err, "clearing users table")

	for i, c := range userCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		err := globalStore.CreateUser(c.user)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		if c.insertErr {
			continue
		}

		var user store.User
		err = rawDB.QueryRow("SELECT username, hashedPassword FROM users WHERE username = $1", c.user.Username).Scan(
			&user.Username, &user.HashedPassword)
		assert.Equal(t, c.getErr, err != nil, caseString)

		assert.Equal(t, c.user, user, caseString)
	}
}

func TestDeleteUser(t *testing.T) {
	_, err := rawDB.Exec("DELETE FROM users")
	assert.Equal(t, nil, err, "clearing users table")

	for i, c := range userCases {
		if c.insertErr {
			continue
		}

		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err := rawDB.Exec("INSERT INTO users VALUES($1, $2)", c.user.Username, c.user.HashedPassword)
		assert.Equal(t, nil, err, caseString)

		err = globalStore.DeleteUser(c.user.Username)
		assert.Equal(t, nil, err, caseString)

		var user store.User
		err = rawDB.QueryRow("SELECT username, hashedPassword FROM users WHERE username = $1", c.user.Username).Scan(
			&user.Username, &user.HashedPassword)
		assert.Equal(t, sql.ErrNoRows, err, caseString)
	}
}

func TestGetTeamsInAlliance(t *testing.T) {
	err := clearTables("reports", "teamsInAlliance", "alliances", "matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	_, err = rawDB.Exec("INSERT INTO events VALUES ($1, $2, $3)", eventKey, "", time.Unix(10000, 0))
	assert.Equal(t, nil, err, "inserting event for testing matches")

	_, err = rawDB.Exec("INSERT INTO matches (key, eventKey, winningAlliance) VALUES ($1, $2, $3)", matchKey, eventKey, "red")
	assert.Equal(t, nil, err, "inserting match for testing alliances")

	var allianceID int
	err = rawDB.QueryRow(
		"INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id",
		matchKey, true, 100).Scan(&allianceID)
	assert.Equal(t, nil, err, "inserting alliance for testing teams")

	for i, c := range teamCases {
		caseString := fmt.Sprintf("case idx: %d\n", i)

		_, err := rawDB.Exec("INSERT INTO teamsInAlliance (number, allianceID, predictedContribution, actualContribution) VALUES($1, $2, $3, $4)", c.test.Number, allianceID, c.test.PredictedContribution, c.test.ActualContribution)
		assert.Equal(t, c.insertErr, err != nil, caseString)

		if c.insertErr {
			continue
		}
	}

	teams, err := globalStore.GetTeamsInAlliance(allianceID)
	assert.Equal(t, nil, err)

	for _, tCase := range teamCases {
		if tCase.insertErr || tCase.getErr {
			continue
		}
		var exists bool
		for _, tGot := range teams {
			tCase.test.AllianceID = tGot.AllianceID
			if tGot == tCase.test {
				exists = true
			}
		}

		assert.Equal(t, true, exists)
	}
}

func TestCreateTeamInAlliance(t *testing.T) {
	err := clearTables("reports", "teamsInAlliance", "alliances", "matches", "events")
	assert.Equal(t, nil, err, "clearing tables")

	_, err = rawDB.Exec("INSERT INTO events VALUES ($1, $2, $3)", eventKey, "", time.Unix(10000, 0))
	assert.Equal(t, nil, err, "inserting event for testing matches")

	_, err = rawDB.Exec("INSERT INTO matches (key, eventKey, winningAlliance) VALUES ($1, $2, $3)", matchKey, eventKey, "red")
	assert.Equal(t, nil, err, "inserting match for testing alliances")

	var allianceID int
	err = rawDB.QueryRow(
		"INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id",
		matchKey, true, 100).Scan(&allianceID)
	assert.Equal(t, nil, err, "inserting alliance for testing teams")

	expectedTeam := store.TeamInAlliance{AllianceID: allianceID, Number: "2733c", PredictedContribution: "Definitly not going to win", ActualContribution: "Yup, didn't win"}
	err = globalStore.CreateTeamInAlliance(allianceID, expectedTeam)
	assert.Equal(t, nil, err, "testing CreateTeamInAlliance failed")

	actualTeam := store.TeamInAlliance{}
	row := rawDB.QueryRow("SELECT number, allianceID, predictedContribution, actualContribution FROM teamsInAlliance WHERE number=$1", "2733c")
	err = row.Scan(&actualTeam.Number, &actualTeam.AllianceID, &actualTeam.PredictedContribution, &actualTeam.ActualContribution)
	assert.Equal(t, nil, err, "testing if CreateTeamInAlliance actually created team failed")
	assert.Equal(t, expectedTeam, actualTeam, "testing if CreateTeamInAlliance created correct team failed")
}

func retrieveCompleteMatch(eventKey string, matchKey string) (store.Match, error) {
	row := rawDB.QueryRow("SELECT key, eventKey, predictedTime, actualTime, winningAlliance FROM matches WHERE eventKey=$1 AND key=$2", eventKey, matchKey)
	var m store.Match
	var predictedTime pq.NullTime
	var actualTime pq.NullTime
	var winningAlliance sql.NullString
	if err := row.Scan(&m.Key, &m.EventKey, &predictedTime, &actualTime, &winningAlliance); err != nil {
		return m, err
	}
	if !predictedTime.Valid {
		m.PredictedTime = time.Time{}
	} else {
		m.PredictedTime = predictedTime.Time.UTC()
	}
	if !actualTime.Valid {
		m.ActualTime = time.Time{}
	} else {
		m.ActualTime = actualTime.Time.UTC()
	}
	if !winningAlliance.Valid {
		m.WinningAlliance = ""
	} else {
		m.WinningAlliance = winningAlliance.String
	}
	redAlliance := store.Alliance{MatchKey: matchKey, IsBlue: false}
	var redID int
	row = rawDB.QueryRow("SELECT id, score FROM alliances WHERE matchKey=$1 AND isBlue=$2", matchKey, false)
	err := row.Scan(&redID, &redAlliance.Score)
	if err != sql.ErrNoRows && err != nil {
		return m, err
	}
	if err == nil {
		redAllianceTeams, err := rawDB.Query("SELECT number, predictedContribution, actualContribution FROM teamsInAlliance WHERE allianceID=$1", redID)
		if err != nil {
			return m, err
		}
		defer redAllianceTeams.Close()
		for redAllianceTeams.Next() {
			var team store.TeamInAlliance
			var predictedContribution sql.NullString
			var actualContribution sql.NullString
			if err := redAllianceTeams.Scan(&team.Number, &predictedContribution, &actualContribution); err != nil {
				return m, err
			}
			if predictedContribution.Valid {
				team.PredictedContribution = predictedContribution.String
			}
			if actualContribution.Valid {
				team.ActualContribution = actualContribution.String
			}
			redAlliance.Teams = append(redAlliance.Teams, team)
		}
		m.RedAlliance = redAlliance
	}

	blueAlliance := store.Alliance{MatchKey: matchKey, IsBlue: true}
	var blueID int
	row = rawDB.QueryRow("SELECT id, score FROM alliances WHERE matchKey=$1 AND isBlue=$2", matchKey, true)
	err = row.Scan(&blueID, &blueAlliance.Score)
	if err != sql.ErrNoRows && err != nil {
		return m, err
	}
	if err == nil {
		blueAllianceTeams, err := rawDB.Query("SELECT number, predictedContribution, actualContribution FROM teamsInAlliance WHERE allianceID=$1", blueID)
		if err != nil {
			return m, err
		}
		defer blueAllianceTeams.Close()
		for blueAllianceTeams.Next() {
			var team store.TeamInAlliance
			var predictedContribution sql.NullString
			var actualContribution sql.NullString
			if err := blueAllianceTeams.Scan(&team.Number, &predictedContribution, &actualContribution); err != nil {
				return m, err
			}
			if predictedContribution.Valid {
				team.PredictedContribution = predictedContribution.String
			}
			if actualContribution.Valid {
				team.ActualContribution = actualContribution.String
			}
			blueAlliance.Teams = append(blueAlliance.Teams, team)
		}
		m.BlueAlliance = blueAlliance
	}
	return m, nil
}
