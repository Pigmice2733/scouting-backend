package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/tbamodified"
)

// Service holds a db for when tba requests were modified.
type Service struct {
	db *sql.DB
}

// New returns a new Service with the given db.
func New(db *sql.DB) tbamodified.Service {
	return &Service{db: db}
}

// Close closes the postgresql db connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// EventsModified returns when the tba data for events was last modified.
func (s *Service) EventsModified() (lastModified string, err error) {
	err = s.db.QueryRow("SELECT lastModified FROM tbaModified WHERE name='events'").Scan(&lastModified)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}
	return
}

// MatchModified returns when the tba data for matches was last modified.
func (s *Service) MatchModified(eventKey string) (lastModified string, err error) {
	err = s.db.QueryRow("SELECT lastModified FROM tbaModified WHERE name=$1", eventKey).Scan(&lastModified)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}
	return
}

// SetEventsModified allows you to set when the tba data for events was last modified.
func (s *Service) SetEventsModified(lastModified string) error {
	if _, err := s.EventsModified(); err == store.ErrNoResults {
		_, err = s.db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES ('events', $1)", lastModified)
		return err
	}

	_, err := s.db.Exec("UPDATE tbaModified SET lastModified=$1 WHERE name='events'", lastModified)
	return err
}

// SetMatchModified allows you to set when the tba data for matches was last modified.
func (s *Service) SetMatchModified(eventKey string, lastModified string) error {
	if _, err := s.MatchModified(eventKey); err == store.ErrNoResults {
		_, err := s.db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES ($1, $2)", eventKey, lastModified)
		return err
	}

	_, err := s.db.Exec("UPDATE tbaModified SET lastModified=$1 WHERE name=$2", lastModified, eventKey)
	return err
}
