package report

import (
	"github.com/Pigmice2733/scouting-backend/internal/analysis"
)

// Report stores information about how a team performed in a match.
type Report struct {
	Reporter string                 `json:"reporter"`
	EventKey string                 `json:"eventKey"`
	MatchKey string                 `json:"matchKey"`
	IsBlue   bool                   `json:"isBlue"`
	Team     string                 `json:"team"`
	Stats    map[string]interface{} `json:"stats"`
}

// Service is a store for reports.
type Service interface {
	Upsert(rep Report) error
	GetReportedOn(eventKey string) ([]string, error)
	GetAllianceReportedOn(eventKey, matchKey string, isBlue bool) ([]string, error)
	GetStatsByEventAndTeam(eventKey, team string) ([]analysis.Data, error)
}
