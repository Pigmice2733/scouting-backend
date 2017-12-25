package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// Service is used for getting information about a match from a postgres database.
type Service struct {
	db *sql.DB
}

// New creates a new match service.
func New(db *sql.DB) match.Service {
	return &Service{db: db}
}

// GetBasicMatches fetches basic information about a match from the postgres database.
func (s *Service) GetBasicMatches(eventKey string) ([]match.BasicMatch, error) {
	var bMatches []match.BasicMatch

	rows, err := s.db.Query("SELECT key, predictedTime, actualTime FROM matches WHERE eventKey = $1", eventKey)
	if err != nil {
		return bMatches, err
	}
	defer rows.Close()

	for rows.Next() {
		var bMatch match.BasicMatch
		if err := rows.Scan(&bMatch.Key, &bMatch.PredictedTime, &bMatch.ActualTime); err != nil {
			return nil, err
		}

		bMatch.EventKey = eventKey
		bMatches = append(bMatches, bMatch)
	}

	return bMatches, rows.Err()
}

// Get gets a full match from the postgres database.
func (s *Service) Get(eventKey, matchKey string, as alliance.Service) (m match.Match, err error) {
	m.Key = matchKey
	m.EventKey = eventKey

	err = s.db.QueryRow("SELECT predictedTime, actualTime, blueWon, redScore, blueScore FROM matches WHERE eventKey = $1 AND key = $2", eventKey, matchKey).Scan(
		&m.PredictedTime, &m.ActualTime, &m.BlueWon, &m.RedScore, &m.BlueScore)
	if err != nil {
		if err == sql.ErrNoRows {
			return m, store.ErrNoResults
		}
		return m, err
	}

	m.RedAlliance, err = as.Get(matchKey, false)
	if err != nil {
		return m, err
	}

	m.BlueAlliance, err = as.Get(matchKey, true)

	return m, err
}

// MassUpsert upserts multiple events in the postgres database.
func (s *Service) MassUpsert(matches []match.Match, as alliance.Service) error {
	stmt, err := s.db.Prepare(`
		INSERT INTO matches (key, eventKey, predictedTime, actualTime, blueWon, redScore, blueScore)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (key)
		DO
			UPDATE
				SET
					eventKey = $2, predictedTime = $3, actualTime = $4,
					blueWon = $5, redScore = $6, blueScore = $7
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, match := range matches {
		if _, err := stmt.Exec(
			match.Key, match.EventKey, match.PredictedTime, match.ActualTime,
			match.BlueWon, match.RedScore, match.BlueScore); err != nil {
			return err
		}
		if err := as.Upsert(match.Key, false, match.RedAlliance); err != nil {
			return err
		}
		if err := as.Upsert(match.Key, true, match.BlueAlliance); err != nil {
			return err
		}
	}

	return nil
}
