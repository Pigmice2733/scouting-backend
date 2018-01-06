package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// Service is used for getting information about an event from a postgres database.
type Service struct {
	db *sql.DB
}

// New creates a new event service.
func New(db *sql.DB) event.Service {
	return &Service{db: db}
}

// GetBasicEvents returns basic event information fetched from the postgres database.
func (s *Service) GetBasicEvents() ([]event.BasicEvent, error) {
	var bEvents []event.BasicEvent

	rows, err := s.db.Query("SELECT key, name, date, shortName, eventType FROM events")
	if err != nil {
		return bEvents, err
	}
	defer rows.Close()

	for rows.Next() {
		var bEvent event.BasicEvent
		if err := rows.Scan(&bEvent.Key, &bEvent.Name, &bEvent.Date, &bEvent.ShortName, &bEvent.EventType); err != nil {
			return nil, err
		}
		bEvents = append(bEvents, bEvent)
	}

	return bEvents, rows.Err()
}

// Get gets a full event from the postgres database.
func (s *Service) Get(key string, ms match.Service) (e event.Event, err error) {
	e.Key = key

	err = s.db.QueryRow("SELECT name, date, shortName, eventType FROM events WHERE key = $1", key).Scan(
		&e.Name, &e.Date, &e.ShortName, &e.EventType)
	if err == sql.ErrNoRows {
		return e, store.ErrNoResults
	} else if err != nil {
		return e, err
	}

	bMatches, err := ms.GetBasicMatches(key)
	e.Matches = bMatches

	return e, err
}

// MassUpsert upserts multiple events in the postgres database.
func (s *Service) MassUpsert(bEvents []event.BasicEvent) error {
	stmt, err := s.db.Prepare(`
		INSERT INTO events (key, name, shortName, date, eventType)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key)
		DO
			UPDATE
				SET name = $2, shortName = $3, date = $4, eventType = $5
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, bEvent := range bEvents {
		if _, err := stmt.Exec(bEvent.Key, bEvent.Name, bEvent.ShortName, bEvent.Date, bEvent.EventType); err != nil {
			return err
		}
	}

	return nil
}
