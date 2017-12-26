package logic

import (
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/analysis"
	"github.com/Pigmice2733/scouting-backend/internal/store/report"
)

// TeamAnalysis holds information about a team, and their analyzed performance.
type TeamAnalysis struct {
	Team  string           `json:"team"`
	Stats analysis.Results `json:"stats"`
}

// EventAnalysis gets information about how all teams at an event performed.
func EventAnalysis(eventKey string, schema analysis.Schema, rs report.Service) ([]TeamAnalysis, error) {
	reportedOn, err := rs.GetReportedOn(eventKey)
	if err != nil {
		return nil, fmt.Errorf("getting teams at event reported on: %v", err)
	}

	return Analyze(eventKey, reportedOn, schema, rs)
}

// AllianceAnalysis gets information about how all teams at a certain event and match of a certain alliance performed.
func AllianceAnalysis(eventKey, matchKey, color string, schema analysis.Schema, rs report.Service) ([]TeamAnalysis, error) {
	reportedOn, err := rs.GetAllianceReportedOn(eventKey, matchKey, color == "blue")
	if err != nil {
		return nil, fmt.Errorf("getting teams on an alliance at a match reported on: %v", err)
	}

	return Analyze(eventKey, reportedOn, schema, rs)
}

// Analyze gets statistics on how a team performed.
func Analyze(eventKey string, teams []string, schema analysis.Schema, rs report.Service) ([]TeamAnalysis, error) {
	teamAnalyses := make([]TeamAnalysis, 0)

	for _, team := range teams {
		stats, err := rs.GetStatsByEventAndTeam(eventKey, team)
		if err != nil {
			return nil, fmt.Errorf("getting stats by event and team: %v", err)
		}

		if len(stats) == 0 {
			return []TeamAnalysis{}, nil
		}

		results, err := analysis.Average(schema, stats...)
		if err != nil {
			return nil, fmt.Errorf("averaging statistics: %v", err)
		}

		teamAnalyses = append(teamAnalyses, TeamAnalysis{Team: team, Stats: results})
	}

	return teamAnalyses, nil
}
