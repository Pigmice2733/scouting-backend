package match

import (
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
)

// Match holds data on a single match, including data on the alliances and their performance
type Match struct {
	Key             string            `json:"key"`
	EventKey        string            `json:"-"`
	PredictedTime   time.Time         `json:"predictedTime,omitempty"`
	ActualTime      time.Time         `json:"actualTime,omitempty"`
	WinningAlliance string            `json:"winningAlliance,omitempty"`
	RedAlliance     alliance.Alliance `json:"redAlliance"`
	BlueAlliance    alliance.Alliance `json:"blueAlliance"`
}

// Service provides an interface for interacting with a store for matches
type Service interface {
	Create(m Match) error
	Get(eventKey, key string) (Match, error)
	GetMatches(eventKey string, as alliance.Service) ([]Match, error)
	UpdateMatches(matches []Match, handlers int, as alliance.Service) []error
	Exists(eventKey, matchKey string) (bool, error)
	Close() error
}
