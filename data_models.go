// data_models.go

package main

import (
	"database/sql"
	"time"
)

type event struct {
	Key  string    `json:"key"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type fullEvent struct {
	Key     string    `json:"key"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date"`
	Matches []match   `json:"matches"`
}

type match struct {
	Key             string `json:"key"`
	EventKey        string `json:"eventKey"`
	Number          int    `json:"number"`
	WinningAlliance string `json:"winningAlliance"`
}

type fullMatch struct {
	Key             string   `json:"key"`
	EventKey        string   `json:"eventKey"`
	Number          int      `json:"number"`
	WinningAlliance string   `json:"winningAlliance"`
	RedAlliance     alliance `json:"redAlliance"`
	BlueAlliance    alliance `json:"blueAlliance"`
}

type alliance struct {
	MatchKey string `json:"matchKey"`
	IsBlue   bool   `json:"isBlue"`
	Score    int    `json:"score"`
	Team1    int    `json:"team1"`
	Team2    int    `json:"team2"`
	Team3    int    `json:"team3"`
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
	rows, err := db.Query("SELECT key, name, date FROM events")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	events := []event{}

	for rows.Next() {
		var e event
		var dateString string
		if err := rows.Scan(&e.Key, &e.Name, &dateString); err != nil {
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
	row := db.QueryRow("SELECT name, date FROM events WHERE key=?", e.Key)

	var dateString string

	if err := row.Scan(&e.Name, &dateString); err != nil {
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
	_, err := db.Exec("INSERT OR IGNORE INTO events(key, name, date) VALUES(?, ?, ?)", e.Key, e.Name, e.Date.Format(time.RFC3339))
	return err
}

func (m *match) getMatch(db *sql.DB) error {
	row := db.QueryRow("SELECT number, winningAlliance FROM matches WHERE eventKey=? AND key=?", m.EventKey, m.Key)

	var winningAlliance sql.NullString
	err := row.Scan(&m.Number, &winningAlliance)
	if err != nil {
		return err
	}

	if !winningAlliance.Valid {
		m.WinningAlliance = ""
	} else {
		m.WinningAlliance = winningAlliance.String
	}
	return nil
}

func getMatches(db *sql.DB, e event) ([]match, error) {
	rows, err := db.Query("SELECT key, eventKey, number, winningAlliance FROM matches WHERE eventKey=?", e.Key)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	matches := []match{}

	for rows.Next() {
		var m match
		var winningAlliance sql.NullString
		if err := rows.Scan(&m.Key, &m.EventKey, &m.Number, &winningAlliance); err != nil {
			return nil, err
		}
		if !winningAlliance.Valid {
			m.WinningAlliance = ""
		} else {
			m.WinningAlliance = winningAlliance.String
		}
		matches = append(matches, m)
	}

	return matches, nil
}

func (m *match) createMatch(db *sql.DB) error {
	_, err := db.Exec("INSERT OR IGNORE INTO matches(key, eventKey, number, winningAlliance) VALUES(?, ?, ?, ?)", m.Key, m.EventKey, m.Number, m.WinningAlliance)
	return err
}

func (a *alliance) getAlliance(db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT id, score, team1, team2, team2 FROM alliances WHERE matchKey=? AND isBlue=?", a.MatchKey, a.IsBlue)

	var allianceID int
	err := row.Scan(&allianceID, &a.Score, &a.Team1, &a.Team2, &a.Team3)
	return allianceID, err
}

func (a *alliance) updateAlliance(db *sql.DB) error {
	_, err := db.Exec("UPDATE alliances SET team1=?, team2=?, team3=? WHERE matchKey=? AND isBlue=?", a.Team1, a.Team2, a.Team3, a.MatchKey, a.IsBlue)
	return err
}

func (a *alliance) createAlliance(db *sql.DB) (int, error) {
	res, err := db.Exec("INSERT INTO alliances(matchKey, score, team1, team2, team3, isBlue) VALUES (?, ?, ?, ?, ?, ?)",
		a.MatchKey, a.Score, a.Team1, a.Team2, a.Team3, a.IsBlue)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
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
