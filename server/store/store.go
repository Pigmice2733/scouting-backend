package store

import (
	"fmt"
	"time"
)

// ErrNoResults is a generic error of sql.ErrNoRows
var ErrNoResults = fmt.Errorf("no results returned")

// Service is a storage interface for event, match, alliance, and report CRUD
type Service interface {
	GetEvents() ([]Event, error)
	UpdateEvents(events []Event) error
	CheckMatchExistence(eventKey string, matchKey string) (bool, error)
	GetEvent(key string) (Event, error)
	CreateEvent(e Event) error
	GetAllMatchData(key string) ([]Match, error)
	GetMatch(eventKey, key string) (Match, error)
	CreateMatch(m Match) error
	GetAlliance(matchKey string, isBlue bool) (Alliance, int, error)
	UpdateAlliance(a Alliance) error
	CreateAlliance(a Alliance) (allianceID int, err error)
	CreateReport(rd ReportData, allianceID int) error
	UpdateReport(rd ReportData, allianceID int) error
	UpdateMatches(matches []Match) error
	EventsModifiedData() (string, error)
	SetEventsModifiedData(lastModified string) error
	SetMatchModifiedData(eventKey string, lastModified string) error
	MatchModifiedData(eventKey string) (string, error)
	GetUser(username string) (User, error)
	GetUsers() ([]User, error)
	CreateUser(User) error
	DeleteUser(username string) error
	GetTeamsInAlliance(allianceID int) ([]TeamInAlliance, error)
	CreateTeamInAlliance(allianceID int, team TeamInAlliance) error
}

// Event holds data from TBA about an event
type Event struct {
	Key  string    `json:"key"`
	Name string    `json:"name"`
	Date time.Time `json:"date,omitempty"`
}

// FullEvent holds an event with match data
type FullEvent struct {
	Key     string    `json:"key"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date,omitempty"`
	Matches []Match   `json:"matches"`
}

// Match holds data on a single match, including data on the alliances and their performance
type Match struct {
	Key             string    `json:"key"`
	EventKey        string    `json:"-"`
	PredictedTime   time.Time `json:"predictedTime,omitempty"`
	ActualTime      time.Time `json:"actualTime,omitempty"`
	WinningAlliance string    `json:"winningAlliance,omitempty"`
	RedAlliance     Alliance  `json:"redAlliance"`
	BlueAlliance    Alliance  `json:"blueAlliance"`
}

// Alliance holds the information on an alliance and its teams
type Alliance struct {
	ID       int              `json:"-"`
	MatchKey string           `json:"-"`
	IsBlue   bool             `json:"-"`
	Score    int              `json:"score"`
	Teams    []TeamInAlliance `json:"teams"`
}

// TeamInAlliance holds the data on how a team performed in an alliance
type TeamInAlliance struct {
	AllianceID            int         `json:"-"`
	Number                string      `json:"number"`
	PredictedContribution interface{} `json:"predictedContribution,omitempty"`
	ActualContribution    interface{} `json:"actualContribution,omitempty"`
}

// AutoReport holds information about the autonomous performance in a match
type AutoReport struct {
	CrossedLine   bool `json:"crossedLine"`
	DeliveredGear bool `json:"deliveredGear"`
	Fuel          int  `json:"fuel"`
}

// TeleopReport holds data about how a team performed in the teleop section of a match
type TeleopReport struct {
	Climbed bool `json:"climbed"`
	Gears   int  `json:"gears"`
	Fuel    int  `json:"fuel"`
}

// ReportData holds information about a team and their the performance in a specific match
type ReportData struct {
	Reporter string       `json:"reporter"`
	Alliance string       `json:"alliance"`
	Team     string       `json:"team"`
	Score    int          `json:"score"`
	Auto     AutoReport   `json:"auto"`
	Teleop   TeleopReport `json:"teleop"`
}

// User holds the credentials for a  user
type User struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
}
