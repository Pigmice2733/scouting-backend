package sqlite3

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Pigmice2733/scouting-backend/store"
)

const eventTableCreationQuery = `
CREATE TABLE IF NOT EXISTS events
(
	key  TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	date TEXT NOT NULL
)`

const matchTableCreationQuery = `
CREATE TABLE IF NOT EXISTS matches
(
	key              TEXT PRIMARY KEY,
	eventKey         TEXT NOT NULL,
	winningAlliance  TEXT,
	FOREIGN KEY(eventKey) REFERENCES events(key)
)`

const allianceTableCreationQuery = `
CREATE TABLE IF NOT EXISTS alliances
(
	id       INTEGER PRIMARY KEY,
	matchKey TEXT    NOT NULL,
	score    INT     NOT NULL,
	team1    INT,
	team2    INT,
	team3    INT,
	isBlue   INT     NOT NULL,
	FOREIGN KEY(matchKey) REFERENCES matches(key)
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

const tbaModifiedTableCreationQuery = `
CREATE TABLE IF NOT EXISTS tbaModified
(
	name         TEXT PRIMARY KEY,
	lastModified TEXT,
	maxAge       TEXT
)
`

type service struct {
	db *sql.DB
}

// NewFromFile returns a new storage service for sqlite3 from a sqlite file
func NewFromFile(dbFileName string) (store.Service, error) {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(eventTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(matchTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(allianceTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(tbaModifiedTableCreationQuery); err != nil {
		log.Fatal(err)
	}

	return New(db), nil
}

// New returns a new storage service for sqlite3
func New(db *sql.DB) store.Service {
	return &service{db}
}

func (s *service) GetEvents() ([]store.Event, error) {
	rows, err := s.db.Query("SELECT key, name, date FROM events")

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNoResults
		}
		return nil, err
	}

	defer rows.Close()

	events := []store.Event{}

	for rows.Next() {
		var e store.Event
		var dateString string
		if err := rows.Scan(&e.Key, &e.Name, &dateString); err != nil {
			return nil, err
		}
		date, err := time.Parse(time.RFC3339, dateString)
		if err != nil {
			return nil, err
		}
		e.Date = date
		events = append(events, e)
	}

	return events, nil
}

func (s *service) GetEvent(e *store.Event) error {
	row := s.db.QueryRow("SELECT name, date FROM events WHERE key=?", e.Key)

	var dateString string

	if err := row.Scan(&e.Name, &dateString); err != nil {
		if err == sql.ErrNoRows {
			return store.ErrNoResults
		}
		return err
	}

	date, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return err
	}
	e.Date = date

	return nil
}

func (s *service) CreateEvent(e store.Event) error {
	_, err := s.db.Exec("INSERT OR IGNORE INTO events(key, name, date) VALUES(?, ?, ?)", e.Key, e.Name, e.Date.Format(time.RFC3339))
	return err
}

func (s *service) GetMatch(m *store.Match) error {
	row := s.db.QueryRow("SELECT winningAlliance FROM matches WHERE eventKey=? AND key=?", m.EventKey, m.Key)

	var winningAlliance sql.NullString
	if err := row.Scan(&winningAlliance); err != nil {
		if err == sql.ErrNoRows {
			return store.ErrNoResults
		}
		return err
	}

	if !winningAlliance.Valid {
		m.WinningAlliance = ""
	} else {
		m.WinningAlliance = winningAlliance.String
	}
	return nil
}

func (s *service) GetMatches(e store.Event) ([]store.Match, error) {
	rows, err := s.db.Query("SELECT key, eventKey, winningAlliance FROM matches WHERE eventKey=?", e.Key)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNoResults
		}
		return nil, err
	}

	defer rows.Close()

	matches := []store.Match{}

	for rows.Next() {
		var m store.Match
		var winningAlliance sql.NullString
		if err := rows.Scan(&m.Key, &m.EventKey, &winningAlliance); err != nil {
			return nil, err
		}
		if !winningAlliance.Valid {
			m.WinningAlliance = ""
		} else {
			m.WinningAlliance = winningAlliance.String
		}
		matches = append(matches, m)
	}

	return matches, nil
}

func (s *service) CreateMatch(m store.Match) error {
	_, err := s.db.Exec("INSERT OR IGNORE INTO matches(key, eventKey, winningAlliance) VALUES(?, ?, ?)", m.Key, m.EventKey, m.WinningAlliance)
	return err
}

func (s *service) GetAlliance(a *store.Alliance) (int, error) {
	row := s.db.QueryRow("SELECT id, score, team1, team2, team2 FROM alliances WHERE matchKey=? AND isBlue=?", a.MatchKey, a.IsBlue)

	var allianceID int
	err := row.Scan(&allianceID, &a.Score, &a.Team1, &a.Team2, &a.Team3)

	if err == sql.ErrNoRows {
		return allianceID, store.ErrNoResults
	}

	return allianceID, err
}

func (s *service) UpdateAlliance(a store.Alliance) error {
	_, err := s.db.Exec("UPDATE alliances SET team1=?, team2=?, team3=? WHERE matchKey=? AND isBlue=?", a.Team1, a.Team2, a.Team3, a.MatchKey, a.IsBlue)
	return err
}

func (s *service) CreateAlliance(a store.Alliance) (int, error) {
	res, err := s.db.Exec("INSERT INTO alliances(matchKey, score, team1, team2, team3, isBlue) VALUES (?, ?, ?, ?, ?, ?)",
		a.MatchKey, a.Score, a.Team1, a.Team2, a.Team3, a.IsBlue)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *service) CreateReport(r store.ReportData, allianceID int) error {
	_, err := s.db.Exec("INSERT INTO reports(allianceID, teamNumber, score, crossedLine, deliveredGear, autoFuel, climbed, gears, teleopFuel) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		allianceID, r.Team, r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel)
	return err
}

func (s *service) UpdateReport(r store.ReportData, allianceID int) error {
	_, err := s.db.Exec("UPDATE reports SET score=?, crossedLine=?, deliveredGear=?, autoFuel=?, climbed=?, gears=?, teleopFuel=? WHERE allianceID=? AND teamNumber=?", r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel, allianceID, r.Team)
	return err
}

func (s *service) UpdateEvents(events []store.Event) error {
	for _, event := range events {
		err := s.CreateEvent(event)
		if err != nil {
			return fmt.Errorf("error processing TBA data '%v' in data '%v'", err.Error(), event)
		}
	}
	return nil
}

func (s *service) UpdateMatches(matches []store.Match) error {
	for _, match := range matches {
		err := s.CreateMatch(match)
		if err != nil {
			return fmt.Errorf("error processing TBA data '%v' in data '%v'", err.Error(), match)
		}
	}
	return nil
}

func (s *service) EventsModifiedData() (string, error) {
	row := s.db.QueryRow("SELECT lastModified FROM tbaModified WHERE name=\"events\"")

	var lastModified string
	if err := row.Scan(&lastModified); err != nil {
		if err == sql.ErrNoRows {
			return "", store.ErrNoResults
		}
		return "", err
	}
	return lastModified, nil
}

func (s *service) SetEventsModifiedData(lastModified string) error {
	_, err := s.EventsModifiedData()
	if err == sql.ErrNoRows {
		_, err = s.db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES (\"events\", ?)", lastModified)
		return err
	}

	_, err = s.db.Exec("UPDATE tbaModified SET lastModified=? WHERE name=\"events\"", lastModified)
	return err
}

func (s *service) SetMatchModifiedData(eventKey string, lastModified string) error {
	if _, err := s.MatchModifiedData(eventKey); err == sql.ErrNoRows {
		_, err := s.db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES (?, ?)", eventKey, lastModified)
		return err
	}

	_, err := s.db.Exec("UPDATE tbaModified SET lastModified=? WHERE name=?", lastModified, eventKey)
	return err
}

func (s *service) MatchModifiedData(eventKey string) (string, error) {
	row := s.db.QueryRow("SELECT lastModified FROM tbaModified WHERE name=?", eventKey)

	var lastModified string
	if err := row.Scan(&lastModified); err != nil {
		return "", err
	}

	return lastModified, nil
}

func (s *service) ensureTableExists(creationQuery string) error {
	_, err := s.db.Exec(creationQuery)
	return err
}

func (s *service) clearTable(t string) error {
	deleteTableContents := fmt.Sprintf("DELETE FROM %v", t)
	if _, err := s.db.Exec(deleteTableContents); err != nil {
		fmt.Println(err)
		return err
	}

	deleteID := fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%v'", t)
	_, err := s.db.Exec(deleteID)
	if err.Error() != "no such table: sqlite_sequence" {
		return err
	}
	return nil
}
