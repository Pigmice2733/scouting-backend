package match

import (
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
)

// BasicMatch holds basic information about a match excluding alliances.
type BasicMatch struct {
	Key           string    `json:"key"`
	EventKey      string    `json:"-"`
	PredictedTime time.Time `json:"predictedTime,omitempty"`
	ActualTime    time.Time `json:"actualTime,omitempty"`
}

// Match holds basic match information and alliance info for the match.
type Match struct {
	BasicMatch
	RedScore     int      `json:"redScore"`
	BlueScore    int      `json:"blueScore"`
	RedAlliance  []string `json:"redAlliance"`
	BlueAlliance []string `json:"blueAlliance"`
}

// Service is a store for matches.
type Service interface {
	GetBasicMatches(eventKey string) ([]BasicMatch, error)
	Get(eventKey, matchKey string, as alliance.Service) (m Match, err error)
	MassUpsert([]Match, alliance.Service) error
}
