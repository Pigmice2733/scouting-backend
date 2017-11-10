package store

import (
	"time"
)

// Service is a storage interface for event, match, alliance, and report CRUD
type Service interface {
	GetEvents() ([]Event, error)
	GetEvent(e *Event) error
	CreateEvent(e Event) error
	GetMatches(e Event) ([]Match, error)
	GetMatch(m *Match) error
	CreateMatch(m Match) error
	GetAlliance(a *Alliance) (int, error)
	UpdateAlliance(a Alliance) error
	CreateAlliance(a Alliance) (int, error)
	CreateReport(rd ReportData, allianceID int) error
	UpdateReport(rd ReportData, allianceID int) error
	UpdateEvents(events []Event) error
	UpdateMatches(matches []Match) error
	EventsModifiedData() (string, error)
	SetEventsModifiedData(lastModified string) error
	SetMatchModifiedData(eventKey string, lastModified string) error
	MatchModifiedData(eventKey string) (string, error)
}

type Event struct {
	Key  string    `json:"key"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type FullEvent struct {
	Key     string    `json:"key"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date"`
	Matches []Match   `json:"matches"`
}

type Match struct {
	Key             string `json:"key"`
	EventKey        string `json:"eventKey"`
	WinningAlliance string `json:"winningAlliance"`
}

type FullMatch struct {
	Key             string   `json:"key"`
	EventKey        string   `json:"eventKey"`
	WinningAlliance string   `json:"winningAlliance"`
	RedAlliance     Alliance `json:"redAlliance"`
	BlueAlliance    Alliance `json:"blueAlliance"`
}

type Alliance struct {
	MatchKey string `json:"matchKey"`
	IsBlue   bool   `json:"isBlue"`
	Score    int    `json:"score"`
	Team1    int    `json:"team1"`
	Team2    int    `json:"team2"`
	Team3    int    `json:"team3"`
}

type AutoReport struct {
	CrossedLine   bool `json:"crossedLine"`
	DeliveredGear bool `json:"deliveredGear"`
	Fuel          int  `json:"fuel"`
}

type TeleopReport struct {
	Climbed bool `json:"climbed"`
	Gears   int  `json:"gears"`
	Fuel    int  `json:"fuel"`
}

type ReportData struct {
	Alliance string       `json:"alliance"`
	Team     int          `json:"team"`
	Score    int          `json:"score"`
	Auto     AutoReport   `json:"auto"`
	Teleop   TeleopReport `json:"teleop"`
}
