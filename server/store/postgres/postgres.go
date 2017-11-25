package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/Pigmice2733/scouting-backend/server/store"
)

const eventTableCreationQuery = `
CREATE TABLE IF NOT EXISTS events
(
	key  TEXT PRIMARY KEY,
	name TEXT              NOT NULL,
	date TIMESTAMPTZ       NOT NULL
)`

const matchTableCreationQuery = `
CREATE TABLE IF NOT EXISTS matches
(
	key                   TEXT PRIMARY KEY,
	eventKey              TEXT              NOT NULL,
	predictedTime         TIMESTAMPTZ,
	actualTime            TIMESTAMPTZ,
	winningAlliance       TEXT,
	FOREIGN KEY(eventKey) REFERENCES events(key)
)`

const allianceTableCreationQuery = `
CREATE TABLE IF NOT EXISTS alliances
(
	id       SERIAL PRIMARY KEY NOT NULL,
	matchKey TEXT               NOT NULL,
	isBlue   BOOLEAN            NOT NULL,
	score    INT                NOT NULL,
	FOREIGN KEY(matchKey) REFERENCES matches(key),
	UNIQUE (matchKey, isBlue)
)
`

// *number is text on purpose, to handle teams like 1540a
const teamInAllianceTableCreationQuery = `
CREATE TABLE IF NOT EXISTS teamsInAlliance
(
	number                TEXT  NOT NULL,
	allianceID            INT   NOT NULL,
	predictedContribution TEXT,
	actualContribution    TEXT,
	FOREIGN KEY(allianceID) REFERENCES alliances(id),
	UNIQUE (number, allianceID)
)
`

const reportTableCreationQuery = `
CREATE TABLE IF NOT EXISTS reports
(
    id            SERIAL PRIMARY KEY NOT NULL,
    reporter      TEXT               NOT NULL,
    allianceID    INT                NOT NULL,
    teamNumber    TEXT               NOT NULL,
    score         INT                NOT NULL,
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
	username       TEXT NOT NULL UNIQUE,
	hashedPassword TEXT NOT NULL
)
`

type service struct {
	db *sql.DB
}

// Options holds information for connecting to a postgres instance
type Options struct {
	User, Pass       string
	Host             string
	Port             int
	DBName           string
	SSLMode          string
	StatementTimeout int
}

func (o Options) connectionInfo() string {
	return fmt.Sprintf("host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode='%s' statement_timeout=%d", o.Host, o.Port, o.User, o.Pass, o.DBName, o.SSLMode, o.StatementTimeout)
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
	if _, err := db.Exec(teamInAllianceTableCreationQuery); err != nil {
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

func (s *service) CreateEvent(e store.Event) error {
	_, err := s.db.Exec("INSERT INTO events(key, name, date) VALUES($1, $2, $3)", e.Key, e.Name, e.Date)
	return err
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

func (s *service) UpdateEvents(events []store.Event) []error {
	errorQueue := make(chan error)

	for _, event := range events {
		go func(e store.Event) {
			_, err := s.db.Exec("INSERT INTO events (key, name, date) VALUES ($1, $2, $3) ON CONFLICT (key) DO UPDATE SET name = $2, date = $3", e.Key, e.Name, e.Date)
			errorQueue <- err
		}(event)
	}

	var errors []error
	for i := 0; i < len(events); i++ {
		if err := <-errorQueue; err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (s *service) CheckMatchExistence(eventKey string, matchKey string) (bool, error) {
	var exists bool
	row := s.db.QueryRow("SELECT EXISTS (SELECT 1 FROM matches WHERE key=$1 AND eventKey=$2)", matchKey, eventKey)
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *service) CreateMatch(m store.Match) error {
	_, err := s.db.Exec("INSERT INTO matches(key, eventKey, predictedTime, actualTime, winningAlliance) VALUES($1, $2, $3, $4, $5)", m.Key, m.EventKey, m.PredictedTime, m.ActualTime, m.WinningAlliance)
	return err
}

func (s *service) GetMatch(eventKey, key string) (store.Match, error) {
	row := s.db.QueryRow("SELECT predictedTime, actualTime, winningAlliance FROM matches WHERE eventKey=$1 AND key=$2", eventKey, key)

	m := store.Match{Key: key, EventKey: eventKey}

	var winningAlliance sql.NullString
	// Golang database/sql doesn't have a NullTime type 🙄
	var predictedTime pq.NullTime
	var actualTime pq.NullTime
	if err := row.Scan(&predictedTime, &actualTime, &winningAlliance); err != nil {
		if err == sql.ErrNoRows {
			return m, store.ErrNoResults
		}
		return m, err
	}

	if !winningAlliance.Valid {
		m.WinningAlliance = ""
	} else {
		m.WinningAlliance = winningAlliance.String
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
	return m, nil
}

func (s *service) GetAllMatchData(eventKey string) ([]store.Match, error) {
	rows, err := s.db.Query("SELECT key, eventKey, predictedTime, actualTime, winningAlliance FROM matches WHERE eventKey=$1", eventKey)

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
		var predictedTime pq.NullTime
		var actualTime pq.NullTime
		var winningAlliance sql.NullString
		if err := rows.Scan(&m.Key, &m.EventKey, &predictedTime, &actualTime, &winningAlliance); err != nil {
			return nil, err
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

		var redAlliance store.Alliance
		var redID int
		if redAlliance, redID, err = s.GetAlliance(m.Key, false); err != nil {
			if err != store.ErrNoResults {
				return nil, err
			}
		} else {
			redAlliance.ID = redID
			redTeams, err := s.GetTeamsInAlliance(redAlliance.ID)
			if err != nil {
				if err != store.ErrNoResults {
					return nil, err
				}
			} else {
				redAlliance.Teams = redTeams
			}
			m.RedAlliance = redAlliance
		}

		var blueAlliance store.Alliance
		var blueID int
		if blueAlliance, blueID, err = s.GetAlliance(m.Key, true); err != nil {
			if err != store.ErrNoResults {
				return nil, err
			}
		} else {
			blueAlliance.ID = blueID
			blueTeams, err := s.GetTeamsInAlliance(blueAlliance.ID)
			if err != nil {
				if err != store.ErrNoResults {
					return nil, err
				}
			} else {
				blueAlliance.Teams = blueTeams
			}
			m.BlueAlliance = blueAlliance
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (s *service) UpdateMatches(matches []store.Match) []error {
	errorQueue := make(chan error)

	for _, match := range matches {
		go func(match store.Match) {
			errorQueue <- s.upsertMatch(match)
		}(match)
	}

	var errors []error
	for i := 0; i < len(matches); i++ {
		if err := <-errorQueue; err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (s *service) CreateAlliance(a store.Alliance) (allianceID int, err error) {
	err = s.db.QueryRow("INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id",
		a.MatchKey, a.IsBlue, a.Score).Scan(&allianceID)
	return allianceID, err
}

func (s *service) GetAlliance(matchKey string, isBlue bool) (store.Alliance, int, error) {
	alliance := store.Alliance{MatchKey: matchKey, IsBlue: isBlue}
	var id int
	row := s.db.QueryRow("SELECT id, score FROM alliances WHERE matchKey=$1 AND isBlue=$2", matchKey, isBlue)
	err := row.Scan(&id, &alliance.Score)
	if err == sql.ErrNoRows {
		return alliance, id, store.ErrNoResults
	}
	alliance.ID = id
	return alliance, id, err
}

// Updates specific alliance
func (s *service) UpdateAlliance(a store.Alliance) error {
	_, err := s.db.Exec("UPDATE alliances SET score=$1 WHERE matchKey=$2 AND isBlue=$3", a.Score, a.MatchKey, a.IsBlue)
	return err
}

func (s *service) CreateReport(r store.ReportData, allianceID int) error {
	_, err := s.db.Exec(
		"INSERT INTO reports(allianceID, reporter, teamNumber, score, crossedLine, deliveredGear, autoFuel, climbed, gears, teleopFuel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		allianceID, r.Reporter, r.Team, r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel)
	return err
}

func (s *service) UpdateReport(r store.ReportData, allianceID int) error {
	_, err := s.db.Exec(
		"UPDATE reports SET reporter = $1, score=$2, crossedLine=$3, deliveredGear=$4, autoFuel=$5, climbed=$6, gears=$7, teleopFuel=$8 WHERE allianceID=$9 AND teamNumber=$10",
		r.Reporter, r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel, allianceID, r.Team)
	return err
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

func (s *service) GetTeamsInAlliance(allianceID int) ([]store.TeamInAlliance, error) {
	var teams []store.TeamInAlliance
	rows, err := s.db.Query("SELECT number, predictedContribution, actualContribution FROM teamsInAlliance WHERE allianceID=$1", allianceID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var team store.TeamInAlliance
		var predictedContribution sql.NullString
		var actualContribution sql.NullString
		if err := rows.Scan(&team.Number, &predictedContribution, &actualContribution); err != nil {
			return nil, err
		}
		if predictedContribution.Valid {
			team.PredictedContribution = predictedContribution.String
		}
		if actualContribution.Valid {
			team.ActualContribution = actualContribution.String
		}
		team.AllianceID = allianceID

		teams = append(teams, team)
	}
	return teams, nil
}

func (s *service) CreateTeamInAlliance(allianceID int, team store.TeamInAlliance) error {
	_, err := s.db.Exec("INSERT INTO teamsInAlliance(number, allianceID, predictedContribution, actualContribution) VALUES ($1, $2, $3, $4)",
		team.Number, allianceID, team.PredictedContribution, team.ActualContribution)
	return err
}

func (s *service) upsertMatch(match store.Match) error {
	transaction, err := s.db.Begin()
	if err != nil {
		return err
	}

	err = upsertOnlyMatch(transaction, match)
	if err != nil {
		transaction.Rollback()
		return err
	}

	redAllianceID, err := upsertAlliance(transaction, match.RedAlliance)
	if err != nil {
		transaction.Rollback()
		return err
	}
	for _, team := range match.RedAlliance.Teams {
		team.AllianceID = redAllianceID
		if err := upsertTeamData(transaction, team); err != nil {
			transaction.Rollback()
			return err
		}
	}

	blueAllianceID, err := upsertAlliance(transaction, match.BlueAlliance)
	if err != nil {
		transaction.Rollback()
		return err
	}
	for _, team := range match.BlueAlliance.Teams {
		team.AllianceID = blueAllianceID
		if err := upsertTeamData(transaction, team); err != nil {
			transaction.Rollback()
			return err
		}
	}

	transaction.Commit()
	return nil
}

// Performs modified upsert - set values are not overwritten with null
func upsertOnlyMatch(tx *sql.Tx, m store.Match) error {
	var winningAllianceData sql.NullString
	row := tx.QueryRow("SELECT winningAlliance FROM matches WHERE eventKey=$1 AND key=$2", m.EventKey, m.Key)
	if err := row.Scan(&winningAllianceData); err != nil {
		if err == sql.ErrNoRows {
			winningAllianceData.Valid = false
		} else {
			return err
		}
	}

	var winner string
	if !winningAllianceData.Valid || m.WinningAlliance != "" {
		winner = m.WinningAlliance
	}

	_, err := tx.Exec("INSERT INTO matches (key, eventKey, predictedTime, actualTime, winningAlliance) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (key) DO UPDATE SET predictedTime = $3, actualTime = $4, winningAlliance = $5", m.Key, m.EventKey, m.PredictedTime, m.ActualTime, winner)
	return err
}

// Performs a modified upsert with an alliance. If value is set
// to null in db but not in struct db is not overwritten. Returns ID of alliance.
func upsertAlliance(tx *sql.Tx, a store.Alliance) (allianceID int, err error) {
	var scoreData sql.NullInt64
	row := tx.QueryRow("SELECT id, score FROM alliances WHERE matchKey=$1 AND isBlue=$2", a.MatchKey, a.IsBlue)
	err = row.Scan(&allianceID, &scoreData)
	if err == sql.ErrNoRows {
		err := tx.QueryRow("INSERT INTO alliances (matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id", a.MatchKey, a.IsBlue, a.Score).Scan(&allianceID)
		return allianceID, err
	} else if err != nil {
		return 0, err
	}
	var score int
	if scoreData.Valid && a.Score == 0 {
		score = int(scoreData.Int64)
	} else {
		score = a.Score
	}
	_, err = tx.Exec("UPDATE alliances SET score=$1 WHERE matchKey=$2 AND isBlue=$3", score, a.MatchKey, a.IsBlue)
	return allianceID, err
}

// Modified upsert - null values won't overwrite set ones
func upsertTeamData(tx *sql.Tx, d store.TeamInAlliance) error {
	var exists bool
	row := tx.QueryRow("SELECT EXISTS (SELECT 1 FROM teamsInAlliance WHERE allianceID=$1 AND number=$2)", d.AllianceID, d.Number)
	err := row.Scan(&exists)
	if err == sql.ErrNoRows || (!exists && err == nil) {
		_, err = tx.Exec("INSERT INTO teamsInAlliance (number, allianceID) VALUES ($1, $2)", d.Number, d.AllianceID)
	}
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
