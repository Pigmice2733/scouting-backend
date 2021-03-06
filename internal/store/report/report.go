package report

import (
	"github.com/Pigmice2733/scouting-backend/internal/analysis"
	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
)

// Report stores information about how a team performed in a match.
type Report struct {
	Reporter string                 `json:"reporter"`
	EventKey string                 `json:"eventKey"`
	MatchKey string                 `json:"matchKey"`
	Team     string                 `json:"team"`
	Notes    *string                `json:"notes"`
	Stats    map[string]interface{} `json:"stats"`
}

// Service is a store for reports.
type Service interface {
	Upsert(rep Report, as alliance.Service) error
	GetReportedOn(eventKey string) ([]string, error)
	GetStatsByEventAndTeam(eventKey, team string) ([]analysis.Data, error)
	GetNotesByEventAndTeam(eventKey, team string) (map[string]string, error)
	GetReportsByEventAndTeam(eventKey, team string) ([]Report, error)
	GetReportsByTeam(team string) ([]Report, error)
	GetReporterStats() (map[string]int, error)
}
