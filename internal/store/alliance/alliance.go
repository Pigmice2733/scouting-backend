package alliance

import (
	"database/sql"
)

// Alliance holds the information on an alliance and its teams
type Alliance struct {
	ID       int    `json:"-"`
	MatchKey string `json:"-"`
	IsBlue   bool   `json:"-"`
	Score    int    `json:"score"`
	Teams    []Team `json:"teams"`
}

// Team holds the information on how a team in an alliance performed.
type Team struct {
	AllianceID            int         `json:"-"`
	Number                string      `json:"number"`
	PredictedContribution interface{} `json:"predictedContribution,omitempty"`
	ActualContribution    interface{} `json:"actualContribution,omitempty"`
}

// Service provides an interface for interacting with a store for alliances.
type Service interface {
	Create(a Alliance) (allianceID int, err error)
	Get(matchKey string, isBlue bool) (Alliance, error)
	Update(a Alliance) error
	GetTeams(allianceID int) ([]Team, error)
	CreateTeam(allianceID int, t Team) error
	Upsert(tx *sql.Tx, a Alliance) (allianceID int, err error)
	UpsertTeam(tx *sql.Tx, t Team) error
	Close() error
}
