// data_models.go

package main

import (
	"database/sql"
	"fmt"
	"time"
)

type event struct {
	ID   uint64    `json:"id"`
	Key  string    `json:"key"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type match struct {
	ID              uint64         `json:"id"`
	EventID         uint64         `json:"eventID"`
	WinningAlliance sql.NullString `json:"winningAlliance"`
	RedAlliance     alliance       `json:"redAlliance"`
	BlueAlliance    alliance       `json:"blueAlliance"`
}

type alliance struct {
	Score       int   `json:"score"`
	TeamNumbers []int `json:"teamNumbers"`
}

type fullEvent struct {
	Event   event   `json:"event"`
	Matches []match `json:"matches"`
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

func getEvent(db *sql.DB, eventID int, e *event) error {
	row := db.QueryRow("SELECT id, key, name, date FROM events WHERE id=?", eventID)

	var dateString string

	if err := row.Scan(&e.ID, &e.Key, &e.Name, &dateString); err != nil {
		return err
	}

	date, err := time.Parse(time.RFC3339, dateString)

	if err != nil {
		return err
	}

	e.Date = date

	return nil
}

func createEvent(db *sql.DB, e event) error {
	_, err := db.Exec("INSERT INTO events(key, name, date) VALUES(?, ?, ?)", e.Key, e.Name, e.Date.Format(time.RFC3339))
	return err
}

func getMatches(db *sql.DB, e event) ([]match, error) {
	rows, err := db.Query("SELECT id, eventID, winningAlliance FROM matches WHERE eventID=?", e.ID)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer rows.Close()

	matches := []match{}

	for rows.Next() {
		var m match
		if err := rows.Scan(&m.ID, &m.EventID, &m.WinningAlliance); err != nil {
			fmt.Println("Match scanning failed " + err.Error())
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, nil
}
