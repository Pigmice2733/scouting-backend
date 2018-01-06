package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
)

// Service is used for getting information about an alliance from a postgres database.
type Service struct {
	db *sql.DB
}

// New creates a new alliance service.
func New(db *sql.DB) alliance.Service {
	return &Service{db: db}
}

// GetColor retrieves the color of the alliance given a matchKey and number.
func (s *Service) GetColor(matchKey string, number string) (isBlue bool, err error) {
	err = s.db.QueryRow("SELECT isBlue FROM alliances WHERE matchKey = $1 AND number = $2", matchKey, number).Scan(&isBlue)
	return
}

// Get gets a certain alliance given a matchKey and whether they were blue or red.
func (s *Service) Get(matchKey string, isBlue bool) (alliance.Alliance, error) {
	alliances := make(alliance.Alliance, 0)

	rows, err := s.db.Query("SELECT number FROM alliances WHERE matchKey = $1 AND isBlue = $2", matchKey, isBlue)
	if err != nil {
		return alliances, err
	}
	defer rows.Close()

	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			return nil, err
		}
		alliances = append(alliances, number)
	}

	return alliances, rows.Err()
}

// Upsert upserts a whole alliance in the postgres database.
func (s *Service) Upsert(matchKey string, isBlue bool, alliance alliance.Alliance) error {
	stmt, err := s.db.Prepare(`
		INSERT INTO alliances (matchKey, isBlue, number)
		VALUES ($1, $2, $3)
		ON CONFLICT (matchKey, number)
		DO
			UPDATE
				SET
					isBlue = $2
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, team := range alliance {
		if _, err := stmt.Exec(matchKey, isBlue, team); err != nil {
			return err
		}
	}

	return nil
}
