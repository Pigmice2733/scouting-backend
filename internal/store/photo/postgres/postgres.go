package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"

	"github.com/Pigmice2733/scouting-backend/internal/store/photo"
)

// Service is used for getting information about a photo from a postgres database.
type Service struct {
	db *sql.DB
}

// New creates a new photo service.
func New(db *sql.DB) photo.Service {
	return &Service{db: db}
}

// Get gets the URL for a team photo from the database.
func (s *Service) Get(team string) (url string, err error) {
	err = s.db.QueryRow("SELECT url FROM photos WHERE team = $1", team).Scan(&url)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}
	return
}

// Create creates a photo URL for a team in the database.
func (s *Service) Create(team, url string) error {
	_, err := s.db.Exec("INSERT INTO photos VALUES ($1, $2)", team, url)
	return err
}
