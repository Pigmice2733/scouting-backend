// data_models.go

package main

import (
	"database/sql"
	"time"
)

type event struct {
	ID   int       `json:"id"`
	Key  string    `json:"key"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type fullEvent struct {
	Event   event   `json:"event"`
	Matches []match `json:"matches"`
}

type match struct {
	ID              int            `json:"id"`
	EventID         int            `json:"eventID"`
	WinningAlliance sql.NullString `json:"winningAlliance"`
}

type fullMatch struct {
	Match        match    `json:"match"`
	RedAlliance  alliance `json:"redAlliance"`
	BlueAlliance alliance `json:"blueAlliance"`
}

type alliance struct {
	MatchID int  `json:"matchID"`
	IsBlue  bool `json:"isBlue"`
	Score   int  `json:"score"`
	Team1   int  `json:"team1"`
	Team2   int  `json:"team2"`
	Team3   int  `json:"team3"`
}

type autoReport struct {
	CrossedLine   bool `json:"crossedLine"`
	DeliveredGear bool `json:"deliveredGear"`
	Fuel          int  `json:"fuel"`
}

type teleopReport struct {
	Climbed bool `json:"climbed"`
	Gears   int  `json:"gears"`
	Fuel    int  `json:"fuel"`
}

type reportData struct {
	Alliance string       `json:"alliance"`
	Team     int          `json:"team"`
	Score    int          `json:"score"`
	Auto     autoReport   `json:"auto"`
	Teleop   teleopReport `json:"teleop"`
}

func getEvents(db *sql.DB) ([]event, error) {
	rows, err := db.Query("SELECT id, key, name, date FROM events")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	events := []event{}

	for rows.Next() {
		var e event
		var dateString string
		if err := rows.Scan(&e.ID, &e.Key, &e.Name, &dateString); err != nil {
			return nil, err
		}
		date, err := time.Parse(time.RFC3339, dateString)
		if err != nil {
			return nil, err
		}
		e.Date = date
		events = append(events, e)
	}

	return events, nil
}

func (e *event) getEvent(db *sql.DB) error {
	row := db.QueryRow("SELECT key, name, date FROM events WHERE id=?", e.ID)

	var dateString string

	if err := row.Scan(&e.Key, &e.Name, &dateString); err != nil {
		return err
	}

	date, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return err
	}
	e.Date = date

	return nil
}

func (e *event) createEvent(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO events(key, name, date) VALUES(?, ?, ?)", e.Key, e.Name, e.Date.Format(time.RFC3339))
	return err
}

func (m *match) getMatch(db *sql.DB) error {
	row := db.QueryRow("SELECT winningAlliance FROM matches WHERE eventID=? AND id=?", m.EventID, m.ID)

	return row.Scan(&m.WinningAlliance)
}

func getMatches(db *sql.DB, e event) ([]match, error) {
	rows, err := db.Query("SELECT id, eventID, winningAlliance FROM matches WHERE eventID=?", e.ID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	matches := []match{}

	for rows.Next() {
		var m match
		if err := rows.Scan(&m.ID, &m.EventID, &m.WinningAlliance); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, nil
}

func (m *match) createMatch(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO matches(eventID, winingAlliance) VALUES(?, ?)", m.EventID, m.WinningAlliance)
	return err
}

func (a *alliance) getAlliance(db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT id, score, team1, team2, team2 FROM alliances WHERE matchID=? AND isBlue=?", a.MatchID, a.IsBlue)

	var allianceID int
	err := row.Scan(&allianceID, &a.Score, &a.Team1, &a.Team2, &a.Team3)
	return allianceID, err
}

func (a *alliance) updateAlliance(db *sql.DB) error {
	_, err := db.Exec("UPDATE alliances SET team1=?, team2=?, team3=? WHERE matchID=? AND isBlue=?", a.Team1, a.Team2, a.Team3, a.MatchID, a.IsBlue)
	return err
}

func (a *alliance) createAlliance(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO alliances(matchID, score, team1, team2, team3, isBlue) VALUES (?, ?, ?, ?, ?, ?)",
		a.MatchID, a.Score, a.Team1, a.Team2, a.Team3, a.IsBlue)
	return err
}

func (r *reportData) createReport(db *sql.DB, allianceID int) error {
	_, err := db.Exec("INSERT INTO reports(allianceID, teamNumber, score, crossedLine, deliveredGear, autoFuel, climbed, gears, teleopFuel) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		allianceID, r.Team, r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel)
	return err
}

func (r *reportData) updateReport(db *sql.DB, allianceID int) error {
	_, err := db.Exec("UPDATE reports SET score=?, crossedLine=?, deliveredGear=?, autoFuel=?, climbed=?, gears=?, teleopFuel=? WHERE allianceID=? AND teamNumber=?", r.Score, r.Auto.CrossedLine, r.Auto.DeliveredGear, r.Auto.Fuel, r.Teleop.Climbed, r.Teleop.Gears, r.Teleop.Fuel, allianceID, r.Team)
	return err
}
