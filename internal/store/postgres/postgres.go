package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	eventPostgres "github.com/Pigmice2733/scouting-backend/internal/store/event/postgres"
	matchPostgres "github.com/Pigmice2733/scouting-backend/internal/store/match/postgres"

	alliancePostgres "github.com/Pigmice2733/scouting-backend/internal/store/alliance/postgres"
	reportPostgres "github.com/Pigmice2733/scouting-backend/internal/store/report/postgres"
	userPostgres "github.com/Pigmice2733/scouting-backend/internal/store/user/postgres"
	// for the postgres sql driver
	_ "github.com/lib/pq"
)

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

// NewFromOptions will connect to a postgresql server with given options
func NewFromOptions(options Options) (*store.Service, error) {
	db, err := sql.Open("postgres", options.connectionInfo())
	if err != nil {
		return nil, err
	}

	return New(db)
}

// New returns a new Service.
func New(db *sql.DB) (*store.Service, error) {
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

	if _, err := db.Exec(userTableCreationQuery); err != nil {
		return nil, err
	}

	return &store.Service{
		Event:    eventPostgres.New(db),
		Match:    matchPostgres.New(db),
		Alliance: alliancePostgres.New(db),
		Report:   reportPostgres.New(db),
		User:     userPostgres.New(db),
	}, nil
}

const eventTableCreationQuery = `
CREATE TABLE IF NOT EXISTS events (
	key TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	shortName TEXT,
	date TIMESTAMPTZ NOT NULL
)
`

const matchTableCreationQuery = `
CREATE TABLE IF NOT EXISTS matches (
	key TEXT PRIMARY KEY,
	eventKey TEXT NOT NULL,
	predictedTime TIMESTAMPTZ,
	actualTime TIMESTAMPTZ,
	blueWon BOOLEAN,
	redScore INTEGER,
	blueScore INTEGER,
	FOREIGN KEY(eventKey) REFERENCES events(key)
)
`

const allianceTableCreationQuery = `
CREATE TABLE IF NOT EXISTS alliances (
	matchKey TEXT NOT NULL,
	isBlue BOOLEAN NOT NULL,
	number TEXT NOT NULL,
	FOREIGN KEY(matchKey) REFERENCES matches(key),
	UNIQUE(matchKey, number)
)
`

/*

TODO: ADD FOREIGN KEY(reporter) REFERENCES users(username),

*/

const reportTableCreationQuery = `
CREATE TABLE IF NOT EXISTS reports (
	reporter TEXT NOT NULL,
	eventKey TEXT NOT NULL,
	matchKey TEXT NOT NULL,
	isBlue BOOLEAN NOT NULL,
	team TEXT NOT NULL,
	stats TEXT NOT NULL,
	UNIQUE(eventKey, matchKey, team),
	FOREIGN KEY(eventKey) REFERENCES events(key),
	FOREIGN KEY(matchKey) REFERENCES matches(key)
)
`

const userTableCreationQuery = `
CREATE TABLE IF NOT EXISTS users (
	username TEXT NOT NULL UNIQUE,
	hashedPassword TEXT NOT NULL
)
`
