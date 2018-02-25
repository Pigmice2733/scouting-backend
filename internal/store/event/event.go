package event

import (
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// BasicEvent holds basic information about an event, not including maches.
type BasicEvent struct {
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	ShortName string    `json:"shortName"`
	EventType int       `json:"eventType"`
	Lat       *float64  `json:"lat,omitempty"`
	Long      *float64  `json:"long,omitempty"`
	Date      time.Time `json:"date"`
}

// Event holds basic information about an event as well as matches.
type Event struct {
	BasicEvent
	Matches []match.BasicMatch `json:"matches"`
}

// Service is a store for events.
type Service interface {
	GetBasicEvents() ([]BasicEvent, error)
	Get(key string, ms match.Service) (Event, error)
	MassUpsert([]BasicEvent) error
}
