package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	eventPostgres "github.com/Pigmice2733/scouting-backend/internal/store/event/postgres"
	matchPostgres "github.com/Pigmice2733/scouting-backend/internal/store/match/postgres"
	"github.com/Pigmice2733/scouting-backend/internal/store/postgres/migrations"

	alliancePostgres "github.com/Pigmice2733/scouting-backend/internal/store/alliance/postgres"
	reportPostgres "github.com/Pigmice2733/scouting-backend/internal/store/report/postgres"
	userPostgres "github.com/Pigmice2733/scouting-backend/internal/store/user/postgres"
	// for the postgres sql driver
	_ "github.com/lib/pq"

	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"
	"github.com/mattes/migrate/source/go-bindata"
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
	driver, err := postgres.WithInstance(db, &postgres.Config{MigrationsTable: "migrations", DatabaseName: "scoutingbackend"})
	if err != nil {
		return nil, fmt.Errorf("1: %v", err)
	}

	s := bindata.Resource(migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return migrations.Asset(name)
		})

	d, err := bindata.WithInstance(s)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("go-bindata", d, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("2: %v", err)
	}

	if err := m.Up(); err != migrate.ErrNoChange && err != nil {
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