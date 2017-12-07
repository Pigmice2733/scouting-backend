package postgres

import (
	"database/sql"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
	"github.com/lib/pq"
)

// Service holds a db for matches.
type Service struct {
	db *sql.DB
}

// New returns a new Service with the given db.
func New(db *sql.DB) match.Service {
	return &Service{db: db}
}

// Close closes the postgresql db connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// Create creates a new match in the postgresql db.
func (s *Service) Create(m match.Match) error {
	_, err := s.db.Exec(
		"INSERT INTO matches(key, eventKey, predictedTime, actualTime, winningAlliance) VALUES($1, $2, $3, $4, $5)",
		m.Key, m.EventKey, m.PredictedTime, m.ActualTime, m.WinningAlliance,
	)
	return err
}

// Get retrieves a match from the postgresql db.
func (s *Service) Get(eventKey, key string) (match.Match, error) {
	row := s.db.QueryRow(
		"SELECT predictedTime, actualTime, winningAlliance FROM matches WHERE eventKey=$1 AND key=$2",
		eventKey, key,
	)

	m := match.Match{Key: key, EventKey: eventKey}

	var winningAlliance sql.NullString
	var predictedTime pq.NullTime
	var actualTime pq.NullTime
	if err := row.Scan(&predictedTime, &actualTime, &winningAlliance); err != nil {
		if err == sql.ErrNoRows {
			return m, store.ErrNoResults
		}
		return m, err
	}

	if !winningAlliance.Valid {
		m.WinningAlliance = ""
	} else {
		m.WinningAlliance = winningAlliance.String
	}

	if !predictedTime.Valid {
		m.PredictedTime = time.Time{}
	} else {
		m.PredictedTime = predictedTime.Time.UTC()
	}

	if !actualTime.Valid {
		m.ActualTime = time.Time{}
	} else {
		m.ActualTime = actualTime.Time.UTC()
	}

	return m, nil
}

// GetMatches retrieves all the matches in the postgresql db for a given event.
func (s *Service) GetMatches(eventKey string, as alliance.Service) ([]match.Match, error) {
	rows, err := s.db.Query("SELECT key, eventKey, predictedTime, actualTime, winningAlliance FROM matches WHERE eventKey=$1", eventKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := []match.Match{}

	for rows.Next() {
		var (
			m               match.Match
			predictedTime   pq.NullTime
			actualTime      pq.NullTime
			winningAlliance sql.NullString
		)

		if err := rows.Scan(&m.Key, &m.EventKey, &predictedTime, &actualTime, &winningAlliance); err != nil {
			return nil, err
		}

		if !predictedTime.Valid {
			m.PredictedTime = time.Time{}
		} else {
			m.PredictedTime = predictedTime.Time.UTC()
		}

		if !actualTime.Valid {
			m.ActualTime = time.Time{}
		} else {
			m.ActualTime = actualTime.Time.UTC()
		}

		if !winningAlliance.Valid {
			m.WinningAlliance = ""
		} else {
			m.WinningAlliance = winningAlliance.String
		}

		var redAlliance alliance.Alliance
		if redAlliance, err = as.Get(m.Key, false); err != nil {
			if err != store.ErrNoResults {
				return nil, err
			}
		} else {
			redTeams, err := as.GetTeams(redAlliance.ID)
			if err != nil {
				if err != store.ErrNoResults {
					return nil, err
				}
			} else {
				redAlliance.Teams = redTeams
			}
			m.RedAlliance = redAlliance
		}

		var blueAlliance alliance.Alliance
		if blueAlliance, err = as.Get(m.Key, true); err != nil {
			if err != store.ErrNoResults {
				return nil, err
			}
		} else {
			blueTeams, err := as.GetTeams(blueAlliance.ID)
			if err != nil {
				if err != store.ErrNoResults {
					return nil, err
				}
			} else {
				blueAlliance.Teams = blueTeams
			}
			m.BlueAlliance = blueAlliance
		}

		matches = append(matches, m)
	}

	return matches, nil
}

// UpdateMatches updates an all given matches in the postgresql db, using as many
// handlers as specified. Using multiple handlers to update the events concurrently
// speeds up the process of updating matches.
func (s *Service) UpdateMatches(matches []match.Match, handlers int, as alliance.Service) []error {
	errorQueue := make(chan error, len(matches))
	matchQueue := make(chan match.Match)

	for i := 0; i < handlers; i++ {
		go func() {
			for m := range matchQueue {
				errorQueue <- s.upsertMatch(m, as)
			}
		}()
	}

	for _, match := range matches {
		matchQueue <- match
	}
	close(matchQueue)

	var errors []error
	for i := 0; i < len(matches); i++ {
		if err := <-errorQueue; err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// Exists returns whether a match is present in the postgresql db or not.
func (s *Service) Exists(eventKey, matchKey string) (exists bool, err error) {
	err = s.db.QueryRow("SELECT EXISTS (SELECT 1 FROM matches WHERE key=$1 AND eventKey=$2)", matchKey, eventKey).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return exists, err
}

func (s *Service) upsertMatch(match match.Match, as alliance.Service) error {
	transaction, err := s.db.Begin()
	if err != nil {
		return err
	}

	err = upsertOnlyMatch(transaction, match)
	if err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}
		return err
	}

	redAllianceID, err := as.Upsert(transaction, match.RedAlliance)
	if err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}
		return err
	}
	for _, team := range match.RedAlliance.Teams {
		team.AllianceID = redAllianceID
		if err := as.UpsertTeam(transaction, team); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}
			return err
		}
	}

	blueAllianceID, err := as.Upsert(transaction, match.BlueAlliance)
	if err != nil {
		if err := transaction.Rollback(); err != nil {
			return err
		}
		return err
	}
	for _, team := range match.BlueAlliance.Teams {
		team.AllianceID = blueAllianceID
		if err := as.UpsertTeam(transaction, team); err != nil {
			if err := transaction.Rollback(); err != nil {
				return err
			}
			return err
		}
	}

	return transaction.Commit()
}

// Performs modified upsert - set values are not overwritten with null
func upsertOnlyMatch(tx *sql.Tx, m match.Match) error {
	var winningAllianceData sql.NullString
	row := tx.QueryRow("SELECT winningAlliance FROM matches WHERE eventKey=$1 AND key=$2", m.EventKey, m.Key)
	if err := row.Scan(&winningAllianceData); err != nil {
		if err == sql.ErrNoRows {
			winningAllianceData.Valid = false
		} else {
			return err
		}
	}

	var winner string
	if !winningAllianceData.Valid || m.WinningAlliance != "" {
		winner = m.WinningAlliance
	}

	_, err := tx.Exec("INSERT INTO matches (key, eventKey, predictedTime, actualTime, winningAlliance) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (key) DO UPDATE SET predictedTime = $3, actualTime = $4, winningAlliance = $5", m.Key, m.EventKey, m.PredictedTime, m.ActualTime, winner)
	return err
}
