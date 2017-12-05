package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
)

// Service holds a db for alliances.
type Service struct {
	db *sql.DB
}

// New returns a new Service with the given db.
func New(db *sql.DB) alliance.Service {
	return &Service{db: db}
}

// Close closes the postgresql db connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// Create creates a new alliance in the postgresql db.
func (s *Service) Create(a alliance.Alliance) (allianceID int, err error) {
	err = s.db.QueryRow("INSERT INTO alliances(matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id",
		a.MatchKey, a.IsBlue, a.Score).Scan(&allianceID)
	return allianceID, err
}

// Get retrieves an alliance from the postgresql db.
func (s *Service) Get(matchKey string, isBlue bool) (alliance.Alliance, error) {
	alliance := alliance.Alliance{MatchKey: matchKey, IsBlue: isBlue}
	err := s.db.QueryRow("SELECT id, score FROM alliances WHERE matchKey=$1 AND isBlue=$2", matchKey, isBlue).Scan(&alliance.ID, &alliance.Score)
	if err == sql.ErrNoRows {
		return alliance, store.ErrNoResults
	}
	return alliance, err
}

// Update updates an alliance in the postgresql db.
func (s *Service) Update(a alliance.Alliance) error {
	_, err := s.db.Exec("UPDATE alliances SET score=$1 WHERE matchKey=$2 AND isBlue=$3", a.Score, a.MatchKey, a.IsBlue)
	return err
}

// GetTeams returns all the teams in an alliance in the postgresql db.
func (s *Service) GetTeams(allianceID int) ([]alliance.Team, error) {
	var teams []alliance.Team
	rows, err := s.db.Query("SELECT number, predictedContribution, actualContribution FROM allianceTeams WHERE allianceID=$1", allianceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var team alliance.Team
		var predictedContribution sql.NullString
		var actualContribution sql.NullString
		if err := rows.Scan(&team.Number, &predictedContribution, &actualContribution); err != nil {
			return nil, err
		}
		if predictedContribution.Valid {
			team.PredictedContribution = predictedContribution.String
		}
		if actualContribution.Valid {
			team.ActualContribution = actualContribution.String
		}
		team.AllianceID = allianceID

		teams = append(teams, team)
	}

	return teams, rows.Err()
}

// CreateTeam will add a team to an alliance in the postgresql db.
func (s *Service) CreateTeam(allianceID int, team alliance.Team) error {
	_, err := s.db.Exec("INSERT INTO allianceTeams(number, allianceID, predictedContribution, actualContribution) VALUES ($1, $2, $3, $4)",
		team.Number, allianceID, team.PredictedContribution, team.ActualContribution)
	return err
}

// Upsert performs a modified upsert with an alliance. If value is set
// to null in db but not in struct db is not overwritten.
func (s *Service) Upsert(tx *sql.Tx, a alliance.Alliance) (allianceID int, err error) {
	var scoreData sql.NullInt64
	row := tx.QueryRow("SELECT id, score FROM alliances WHERE matchKey=$1 AND isBlue=$2", a.MatchKey, a.IsBlue)
	err = row.Scan(&allianceID, &scoreData)
	if err == sql.ErrNoRows {
		err := tx.QueryRow("INSERT INTO alliances (matchKey, isBlue, score) VALUES ($1, $2, $3) RETURNING id", a.MatchKey, a.IsBlue, a.Score).Scan(&allianceID)
		return allianceID, err
	} else if err != nil {
		return 0, err
	}
	var score int
	if scoreData.Valid && a.Score == 0 {
		score = int(scoreData.Int64)
	} else {
		score = a.Score
	}
	_, err = tx.Exec("UPDATE alliances SET score=$1 WHERE matchKey=$2 AND isBlue=$3", score, a.MatchKey, a.IsBlue)
	return allianceID, err
}

// UpsertTeam performs a modified upsert: null values won't overwrite set ones.
func (s *Service) UpsertTeam(tx *sql.Tx, t alliance.Team) error {
	var exists bool
	err := tx.QueryRow("SELECT EXISTS (SELECT 1 FROM allianceTeams WHERE allianceID=$1 AND number=$2)", t.AllianceID, t.Number).Scan(&exists)
	if err == sql.ErrNoRows || (!exists && err == nil) {
		_, err = tx.Exec("INSERT INTO allianceTeams (number, allianceID) VALUES ($1, $2)", t.Number, t.AllianceID)
	}
	return err
}
