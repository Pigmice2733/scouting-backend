package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Pigmice2733/scouting-backend/server/store"
	// Register postgres driver
	_ "github.com/lib/pq"
)

const eventTableCreationQuery = `
CREATE TABLE IF NOT EXISTS events
(
	key  TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	date TIMESTAMP NOT NULL
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
	id       SERIAL PRIMARY KEY NOT NULL,
	matchKey TEXT    NOT NULL,
	score    INT     NOT NULL,
	team1    INT,
	team2    INT,
	team3    INT,
	isBlue   BOOLEAN     NOT NULL,
	FOREIGN KEY(matchKey) REFERENCES matches(key)
)
`

const reportTableCreationQuery = `
CREATE TABLE IF NOT EXISTS reports
(
    id            SERIAL PRIMARY KEY NOT NULL,
    allianceID    INT     NOT NULL,
    teamNumber    INT     NOT NULL,
    score         INT     NOT NULL,
    crossedLine   BOOLEAN,
    deliveredGear BOOLEAN,
    autoFuel      INT,
    climbed       BOOLEAN,
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

const usersTableCreationQuery = `
CREATE TABLE IF NOT EXISTS users
(
	username TEXT NOT NULL UNIQUE,
	hashedPassword TEXT NOT NULL
)
`

type service struct {
	db *sql.DB
}

// Options holds information for connecting to a postgres instance
type Options struct {
	User, Pass string
	Host       string
	Port       int
	DBName     string
	SSLMode    string
}

func (o Options) connectionInfo() string {
	return fmt.Sprintf("host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode='%s'", o.Host, o.Port, o.User, o.Pass, o.DBName, o.SSLMode)
}

// NewFromOptions creates a new storage service from provided connection options
func NewFromOptions(options Options) (store.Service, error) {
	db, err := sql.Open("postgres", options.connectionInfo())

	if _, err := db.Exec(eventTableCreationQuery); err != nil {
		return nil, err
	}
	if _, err := db.Exec(matchTableCreationQuery); err != nil {
		return nil, err
	}
	if _, err := db.Exec(allianceTableCreationQuery); err != nil {
		return nil, err
	}
	if _, err := db.Exec(reportTableCreationQuery); err != nil {
		return nil, err
	}
	if _, err := db.Exec(tbaModifiedTableCreationQuery); err != nil {
		return nil, err
	}
	if _, err := db.Exec(usersTableCreationQuery); err != nil {
		return nil, err
	}

	return &service{db}, err
}

// New returns a new storage service for postgres
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
		if err := rows.Scan(&e.Key, &e.Name, &e.Date); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func (s *service) GetEvent(key string) (store.Event, error) {
	row := s.db.QueryRow("SELECT name, date FROM events WHERE key=$1", key)

	e := store.Event{Key: key}

	if err := row.Scan(&e.Name, &e.Date); err != nil {
		if err == sql.ErrNoRows {
			return e, store.ErrNoResults
		}
		return e, err
	}

	return e, nil
}

func (s *service) CreateEvent(e store.Event) error {
	_, err := s.db.Exec("INSERT INTO events(key, name, date) VALUES($1, $2, $3)", e.Key, e.Name, e.Date)
	return err
}

func (s *service) GetMatch(eventKey, key string) (store.Match, error) {
	row := s.db.QueryRow("SELECT winningAlliance FROM matches WHERE eventKey=$1 AND key=$2", eventKey, key)

	var winningAlliance sql.NullString
	e := store.Match{EventKey: eventKey, Key: key}

	if err := row.Scan(&winningAlliance); err != nil {
		if err == sql.ErrNoRows {
			return e, store.ErrNoResults
		}
		return e, err
	}

	if !winningAlliance.Valid {
		e.WinningAlliance = ""
	} else {
		e.WinningAlliance = winningAlliance.String
	}

	return e, nil
}

func (s *service) GetMatches(key string) ([]store.Match, error) {
	rows, err := s.db.Query("SELECT key, eventKey, winningAlliance FROM matches WHERE eventKey=$1", key)

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
	_, err := s.db.Exec("INSERT INTO matches(key, eventKey, winningAlliance) VALUES($1, $2, $3)", m.Key, m.EventKey, m.WinningAlliance)
	return err
}

func (s *service) GetAlliance(matchKey string, isBlue bool) (store.Alliance, int, error) {
	row := s.db.QueryRow("SELECT id, score, team1, team2, team3 FROM alliances WHERE matchKey=$1 AND isBlue=$2", matchKey, isBlue)

	var allianceID int
	a := store.Alliance{MatchKey: matchKey, IsBlue: isBlue}

	err := row.Scan(&allianceID, &a.Score, &a.Team1, &a.Team2, &a.Team3)

	if err == sql.ErrNoRows {
		return a, allianceID, store.ErrNoResults
	}

	return a, allianceID, err
}

func (s *service) UpdateAlliance(a store.Alliance) error {
	_, err := s.db.Exec("UPDATE alliances SET score = $1, team1=$2, team2=$3, team3=$4 WHERE matchKey=$5 AND isBlue=$6", a.Score, a.Team1, a.Team2, a.Team3, a.MatchKey, a.IsBlue)
	return err
}

func (s *service) CreateAlliance(a store.Alliance) (id int, err error) {
	err = s.db.QueryRow("INSERT INTO alliances(matchKey, score, team1, team2, team3, isBlue) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		a.MatchKey, a.Score, a.Team1, a.Team2, a.Team3, a.IsBlue).Scan(&id)
	return id, err
}

func (s *service) CreateReport(r store.ReportData, allianceID int) error {
	_, err := s.db.Exec("INSERT INTO reports(allianceID, teamNumber, score, crossedLine, deliveredGear, autoFuel, climbed, gears, teleopFuel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		allianceID, r.Team, r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel)
	return err
}

func (s *service) UpdateReport(r store.ReportData, allianceID int) error {
	_, err := s.db.Exec("UPDATE reports SET score=$1, crossedLine=$2, deliveredGear=$3, autoFuel=$4, climbed=$5, gears=$6, teleopFuel=$7 WHERE allianceID=$8 AND teamNumber=$9", r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel, allianceID, r.Team)
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
		_, err = s.db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES (\"events\", $1)", lastModified)
		return err
	}

	_, err = s.db.Exec("UPDATE tbaModified SET lastModified=$1 WHERE name=\"events\"", lastModified)
	return err
}

func (s *service) SetMatchModifiedData(eventKey string, lastModified string) error {
	if _, err := s.MatchModifiedData(eventKey); err == sql.ErrNoRows {
		_, err := s.db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES ($1, $2)", eventKey, lastModified)
		return err
	}

	_, err := s.db.Exec("UPDATE tbaModified SET lastModified=$1 WHERE name=$2", lastModified, eventKey)
	return err
}

func (s *service) MatchModifiedData(eventKey string) (string, error) {
	row := s.db.QueryRow("SELECT lastModified FROM tbaModified WHERE name=$1", eventKey)

	var lastModified string
	if err := row.Scan(&lastModified); err != nil {
		return "", err
	}

	return lastModified, nil
}

func (s *service) GetUser(username string) (store.User, error) {
	var user store.User
	err := s.db.QueryRow("SELECT username, hashedPassword FROM users WHERE username = $1", username).Scan(&user.Username, &user.HashedPassword)
	if err == sql.ErrNoRows {
		return user, store.ErrNoResults
	}
	return user, err
}

func (s *service) GetUsers() ([]store.User, error) {
	var users []store.User

	rows, err := s.db.Query("SELECT username, hashedPassword FROM users")
	if err != nil {
		if err == sql.ErrNoRows {
			return users, store.ErrNoResults
		}
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user store.User
		if err := rows.Scan(&user.Username, &user.HashedPassword); err != nil {
			return users, err
		}
		users = append(users, user)
	}

	err = rows.Err()

	return users, err
}

func (s *service) CreateUser(user store.User) error {
	_, err := s.db.Exec("INSERT INTO users VALUES ($1, $2)", user.Username, user.HashedPassword)
	return err
}

func (s *service) DeleteUser(username string) error {
	_, err := s.db.Exec("DELETE FROM users WHERE username = $1", username)
	return err
}

func (s *service) ensureTableExists(creationQuery string) error {
	_, err := s.db.Exec(creationQuery)
	return err
}

func (s *service) clearTable(t string) error {
	_, err := s.db.Exec(fmt.Sprintf("DELETE FROM %s", t))
	return err
}
