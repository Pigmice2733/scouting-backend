package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store/report"
)

// Service holds a db for reports.
type Service struct {
	db *sql.DB
}

// New returns a new Service with the given db.
func New(db *sql.DB) report.Service {
	return &Service{db: db}
}

// Close closes the postgresql db connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// Create adds a new report to the postgresql db.
func (s *Service) Create(r report.Report, allianceID int) error {
	_, err := s.db.Exec(
		"INSERT INTO reports(allianceID, reporter, teamNumber, score, crossedLine, deliveredGear, autoFuel, climbed, gears, teleopFuel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		allianceID, r.Reporter, r.Team, r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel)
	return err
}

// Update updates an existing report in the postgresql db.
func (s *Service) Update(r report.Report, allianceID int) error {
	_, err := s.db.Exec(
		"UPDATE reports SET reporter = $1, score=$2, crossedLine=$3, deliveredGear=$4, autoFuel=$5, climbed=$6, gears=$7, teleopFuel=$8 WHERE allianceID=$9 AND teamNumber=$10",
		r.Reporter, r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel, allianceID, r.Team)
	return err
}
