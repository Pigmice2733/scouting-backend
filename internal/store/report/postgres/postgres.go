package postgres

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"github.com/Pigmice2733/scouting-backend/internal/analysis"
	"github.com/Pigmice2733/scouting-backend/internal/store/report"
)

// Service is used for getting information about a report from a postgres database.
type Service struct {
	db *sql.DB
}

// New creates a new report service.
func New(db *sql.DB) report.Service {
	return &Service{db: db}
}

// Upsert upserts (creates if the resource doesn't exist, otherwise updates) a report into the postgresql database.
func (s *Service) Upsert(rep report.Report) error {
	stats := new(bytes.Buffer)
	if err := json.NewEncoder(stats).Encode(rep.Stats); err != nil {
		return err
	}

	_, err := s.db.Exec(`
		INSERT INTO reports (reporter, isBlue, team, stats, eventKey, matchKey)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (eventKey, matchKey, team)
		DO
			UPDATE
				SET reporter = $1, isBlue = $2, team = $3, stats = $4
	`, rep.Reporter, rep.IsBlue, rep.Team, stats.String(), rep.EventKey, rep.MatchKey)
	return err
}

// GetReportedOn gets all teams that have been reported on at an event.
func (s *Service) GetReportedOn(eventKey string) (reportedOn []string, err error) {
	rows, err := s.db.Query("SELECT DISTINCT team FROM reports WHERE eventKey = $1", eventKey)
	if err != nil {
		return reportedOn, err
	}
	defer rows.Close()

	for rows.Next() {
		var team string
		if err := rows.Scan(&team); err != nil {
			return reportedOn, err
		}

		reportedOn = append(reportedOn, team)
	}

	return reportedOn, rows.Err()
}

// GetAllianceReportedOn gets all teams reported on at an event at a match of a certain color (blue or red).
func (s *Service) GetAllianceReportedOn(eventKey, matchKey string, isBlue bool) (reportedOn []string, err error) {
	rows, err := s.db.Query("SELECT team FROM reports WHERE eventKey = $1 AND matchKey = $2 AND isBlue = $3", eventKey, matchKey, isBlue)
	if err != nil {
		return reportedOn, err
	}
	defer rows.Close()

	for rows.Next() {
		var team string
		if err := rows.Scan(&team); err != nil {
			return reportedOn, err
		}

		reportedOn = append(reportedOn, team)
	}

	return reportedOn, rows.Err()
}

// GetStatsByEventAndTeam gets all statistics from reports of a certain team at a certain event.
func (s *Service) GetStatsByEventAndTeam(eventKey, team string) ([]analysis.Data, error) {
	rows, err := s.db.Query("SELECT stats FROM reports WHERE eventKey = $1 AND team = $2", eventKey, team)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]analysis.Data, 0)

	for rows.Next() {
		stat := make(analysis.Data)

		var statsStr string
		if err := rows.Scan(&statsStr); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(statsStr), &stat); err != nil {
			return nil, err
		}

		stats = append(stats, stat)
	}

	return stats, rows.Err()
}
