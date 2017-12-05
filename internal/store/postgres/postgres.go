package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	postgresAlliance "github.com/Pigmice2733/scouting-backend/internal/store/alliance/postgres"
	postgresEvent "github.com/Pigmice2733/scouting-backend/internal/store/event/postgres"
	postgresMatch "github.com/Pigmice2733/scouting-backend/internal/store/match/postgres"
	postgresReport "github.com/Pigmice2733/scouting-backend/internal/store/report/postgres"
	postgresTBAModified "github.com/Pigmice2733/scouting-backend/internal/store/tbamodified/postgres"
	postgresUser "github.com/Pigmice2733/scouting-backend/internal/store/user/postgres"
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
	tableCreationQueries := []string{
		eventTableCreationQuery,
		matchTableCreationQuery,
		allianceTableCreationQuery,
		allianceTeamsTableCreationQuery,
		reportTableCreationQuery,
		tbaModifiedTableCreationQuery,
		usersTableCreationQuery,
	}

	for _, tableCreationQuery := range tableCreationQueries {
		if _, err := db.Exec(tableCreationQuery); err != nil {
			return nil, err
		}
	}

	return &store.Service{
		Alliance:    postgresAlliance.New(db),
		Event:       postgresEvent.New(db),
		Match:       postgresMatch.New(db),
		Report:      postgresReport.New(db),
		TBAModified: postgresTBAModified.New(db),
		User:        postgresUser.New(db),
	}, nil
}

const eventTableCreationQuery = `
CREATE TABLE IF NOT EXISTS events (
	key  TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	date TIMESTAMPTZ NOT NULL
)
`

const matchTableCreationQuery = `
CREATE TABLE IF NOT EXISTS matches (
	key             TEXT PRIMARY KEY,
	eventKey        TEXT NOT NULL,
	predictedTime   TIMESTAMPTZ,
	actualTime      TIMESTAMPTZ,
	winningAlliance TEXT,
	FOREIGN KEY(eventKey) REFERENCES events(key)
)
`

const allianceTableCreationQuery = `
CREATE TABLE IF NOT EXISTS alliances (
	id       SERIAL PRIMARY KEY NOT NULL,
	matchKey TEXT NOT NULL,
	isBlue   BOOLEAN NOT NULL,
	score    INT NOT NULL,
	FOREIGN KEY(matchKey) REFERENCES matches(key),
	UNIQUE (matchKey, isBlue)
)
`

const allianceTeamsTableCreationQuery = `
CREATE TABLE IF NOT EXISTS allianceTeams
(
	number                TEXT NOT NULL,
	allianceID            INT NOT NULL,
	predictedContribution TEXT,
	actualContribution    TEXT,
	FOREIGN KEY(allianceID) REFERENCES alliances(id),
	UNIQUE (number, allianceID)
)
`

const reportTableCreationQuery = `
CREATE TABLE IF NOT EXISTS reports
(
    id            SERIAL PRIMARY KEY,
    reporter      TEXT NOT NULL,
    allianceID    INT NOT NULL,
    teamNumber    TEXT NOT NULL,
    score         INT NOT NULL,
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
	lastModified TEXT NOT NULL
)
`

const usersTableCreationQuery = `
CREATE TABLE IF NOT EXISTS users
(
	username       TEXT NOT NULL UNIQUE,
	hashedPassword TEXT NOT NULL
)
`
