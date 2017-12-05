package event

import (
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// Event holds data from TBA about an event.
type Event struct {
	Key     string        `json:"key"`
	Name    string        `json:"name"`
	Date    time.Time     `json:"date"`
	Matches []match.Match `json:"matches,omitempty"`
}

// Service provides an interface for interacting with a store for events.
type Service interface {
	Create(e Event) error
	Get(key string) (Event, error)
	GetEvents() ([]Event, error)
	UpdateEvents(events []Event, handlers int) []error
	Close() error
}
