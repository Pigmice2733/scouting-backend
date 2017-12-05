package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/event"
)

// Service holds a db for events.
type Service struct {
	db *sql.DB
}

// New returns a new Service with the given db.
func New(db *sql.DB) event.Service {
	return &Service{db: db}
}

// Close closes the postgresql db connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// Create creates a new event in the postgresql db.
func (s *Service) Create(e event.Event) error {
	_, err := s.db.Exec("INSERT INTO events(key, name, date) VALUES($1, $2, $3)", e.Key, e.Name, e.Date)
	return err
}

// Get retrieves an event from the postgresql db.
func (s *Service) Get(key string) (event.Event, error) {
	e := event.Event{Key: key}
	err := s.db.QueryRow("SELECT name, date FROM events WHERE key=$1", key).Scan(&e.Name, &e.Date)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}

	return e, err
}

// GetEvents returns all the events in the postgresql db.
func (s *Service) GetEvents() ([]event.Event, error) {
	rows, err := s.db.Query("SELECT key, name, date FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []event.Event{}

	for rows.Next() {
		var e event.Event
		if err := rows.Scan(&e.Key, &e.Name, &e.Date); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// UpdateEvents updates an all given events in the postgresql db, using as many
// handlers as specified. Using multiple handlers to update the events concurrently
// speeds up the process of updating events.
func (s *Service) UpdateEvents(events []event.Event, handlers int) []error {
	errorQueue := make(chan error, len(events))
	eventQueue := make(chan event.Event)

	for i := 0; i < handlers; i++ {
		go func() {
			for e := range eventQueue {
				_, err := s.db.Exec("INSERT INTO events (key, name, date) VALUES ($1, $2, $3) ON CONFLICT (key) DO UPDATE SET name = $2, date = $3", e.Key, e.Name, e.Date)
				errorQueue <- err
			}
		}()
	}

	for _, event := range events {
		eventQueue <- event
	}
	close(eventQueue)

	var errors []error
	for i := 0; i < len(events); i++ {
		if err := <-errorQueue; err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
