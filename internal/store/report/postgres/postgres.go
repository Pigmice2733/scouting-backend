package postgres

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"github.com/Pigmice2733/scouting-backend/internal/analysis"
	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
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
func (s *Service) Upsert(rep report.Report, as alliance.Service) error {
	stats := new(bytes.Buffer)
	if err := json.NewEncoder(stats).Encode(rep.Stats); err != nil {
		return err
	}

	_, err := s.db.Exec(`
		INSERT INTO reports (reporter, team, stats, notes, eventKey, matchKey)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (eventKey, matchKey, team)
		DO
			UPDATE
				SET reporter = $1, team = $2, stats = $3, notes = $4
	`, rep.Reporter, rep.Team, stats.String(), rep.Notes, rep.EventKey, rep.MatchKey)

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

// GetNotesByEventAndTeam gets all notes from reports of a certain team at a certain event.
func (s *Service) GetNotesByEventAndTeam(eventKey, team string) (map[string]string, error) {
	rows, err := s.db.Query("SELECT matchKey, notes FROM reports WHERE eventKey = $1 AND team = $2 AND notes IS NOT NULL", eventKey, team)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make(map[string]string)

	for rows.Next() {
		var note string
		var matchKey string

		if err := rows.Scan(&matchKey, &note); err != nil {
			return nil, err
		}

		notes[matchKey] = note
	}

	return notes, rows.Err()
}

// GetReportsByEventAndTeam gets all reports on a certain team at a certain event.
func (s *Service) GetReportsByEventAndTeam(eventKey, team string) ([]report.Report, error) {
	rows, err := s.db.Query("SELECT reporter, matchKey, stats, notes FROM reports WHERE eventKey = $1 AND team = $2", eventKey, team)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []report.Report

	for rows.Next() {
		var rep report.Report
		var statsStr string

		if err := rows.Scan(&rep.Reporter, &rep.MatchKey, &statsStr, &rep.Notes); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(statsStr), &rep.Stats); err != nil {
			return nil, err
		}

		rep.EventKey = eventKey
		rep.Team = team

		reports = append(reports, rep)
	}

	return reports, rows.Err()
}

// GetReportsByTeam gets all reports on a certain team from all events.
func (s *Service) GetReportsByTeam(team string) ([]report.Report, error) {
	rows, err := s.db.Query("SELECT reporter, eventKey, matchKey, stats, notes FROM reports WHERE team = $1", team)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []report.Report

	for rows.Next() {
		var rep report.Report
		var statsStr string

		if err := rows.Scan(&rep.Reporter, &rep.EventKey, &rep.MatchKey, &statsStr, &rep.Notes); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(statsStr), &rep.Stats); err != nil {
			return nil, err
		}

		rep.Team = team

		reports = append(reports, rep)
	}

	return reports, rows.Err()
}

// GetReporterStats gets a map of all reporters to the amount of reports they have submitted.
func (s *Service) GetReporterStats() (map[string]int, error) {
	rows, err := s.db.Query("SELECT reporter, COUNT(reporter) FROM reports GROUP BY reporter")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)

	for rows.Next() {
		var reporter string
		var count int

		if err := rows.Scan(&reporter, &count); err != nil {
			return nil, err
		}

		stats[reporter] = count
	}

	return stats, rows.Err()
}
