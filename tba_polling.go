package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Pigmice2733/scouting-backend/logger"
)

func pollTBAEvents(db *sql.DB, logger logger.Service, tbaAPI string, apikey string, year string) error {
	type tbaEvent struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Date string `json:"start_date"`
	}
	var tbaEvents []tbaEvent

	eventsEndpoint := fmt.Sprintf("%s/events/%s", tbaAPI, year)

	req, err := http.NewRequest("GET", eventsEndpoint, nil)
	if err != nil {
		return fmt.Errorf("TBA polling failed with error %s", err)
	}

	lastModified, err := eventsModifiedData(db)
	if err == nil {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", apikey)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("TBA polling request failed with error %s", err)
	}

	lastModified = response.Header.Get("Last-Modified")
	if lastModified != "" {
		setEventsModifiedData(db, lastModified)
	}

	responseCode := response.StatusCode

	if responseCode == http.StatusNotModified {
		logger.Debugf("TBA event data not modified")
		return nil
	} else if responseCode != http.StatusOK {
		return fmt.Errorf("TBA polling request failed with status %d", responseCode)
	}

	eventData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading TBA response")
	}
	json.Unmarshal(eventData, &tbaEvents)

	var events []event

	for _, tbaEvent := range tbaEvents {
		date, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			logger.Debugf("Error TBA time data %v", err.Error())
			continue
		}
		newEvent := event{
			Key:  tbaEvent.Key,
			Name: tbaEvent.Name,
			Date: date,
		}
		events = append(events, newEvent)
	}

	err = updateEvents(db, events)
	if err != nil {
		return err
	}

	logger.Infof("Polled TBA...")
	return nil
}

func pollTBAMatches(db *sql.DB, tbaAPI string, apikey string, eventKey string) ([]match, error) {
	type tbaMatch struct {
		Key    string `json:"key"`
	}
	var tbaMatches []tbaMatch

	matchesEndpoint := fmt.Sprintf("%s/event/%s/matches", tbaAPI, eventKey)

	req, err := http.NewRequest("GET", matchesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	lastModified, err := matchModifiedData(db, eventKey)
	if err == nil {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", apikey)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	lastModified = response.Header.Get("Last-Modified")
	if lastModified != "" {
		setMatchModifiedData(db, lastModified, eventKey)
	}

	responseCode := response.StatusCode

	if responseCode == http.StatusNotModified {
		return nil, nil
	} else if responseCode != http.StatusOK {
		return nil, fmt.Errorf("TBA polling request failed with status %v", responseCode)
	}

	matchData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(matchData, &tbaMatches)

	var matches []match

	for _, tbaMatch := range tbaMatches {
		newMatch := match{
			Key:      tbaMatch.Key,
			EventKey: eventKey,
		}
		matches = append(matches, newMatch)
	}

	err = updateMatches(db, matches)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func updateEvents(db *sql.DB, events []event) error {
	for _, event := range events {
		err := event.createEvent(db)
		if err != nil {
			return fmt.Errorf("Error processing TBA data '%v' in data '%v'", err.Error(), event)
		}
	}
	return nil
}

func updateMatches(db *sql.DB, matches []match) error {
	for _, match := range matches {
		err := match.createMatch(db)
		if err != nil {
			return fmt.Errorf("Error processing TBA data '%v' in data '%v'", err.Error(), match)
		}
	}
	return nil
}

func eventsModifiedData(db *sql.DB) (string, error) {
	row := db.QueryRow("SELECT lastModified FROM tbaModified WHERE name=\"events\"")

	var lastModified string
	if err := row.Scan(&lastModified); err != nil {
		return "", err
	}
	return lastModified, nil
}

func setEventsModifiedData(db *sql.DB, lastModified string) {
	_, err := eventsModifiedData(db)
	if err == sql.ErrNoRows {
		db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES (\"events\", ?)", lastModified)
	} else {
		db.Exec("UPDATE tbaModified SET lastModified=? WHERE name=\"events\"", lastModified)
	}
}

func matchModifiedData(db *sql.DB, eventKey string) (string, error) {
	row := db.QueryRow("SELECT lastModified FROM tbaModified WHERE name=?", eventKey)

	var lastModified string
	if err := row.Scan(&lastModified); err != nil {
		return "", err
	}
	return lastModified, nil
}

func setMatchModifiedData(db *sql.DB, eventKey string, lastModified string) {
	_, err := matchModifiedData(db, eventKey)
	if err == sql.ErrNoRows {
		db.Exec("INSERT INTO tbaModified(name, lastModified) VALUES (?, ?)", eventKey, lastModified)
	} else {
		db.Exec("UPDATE tbaModified SET lastModified=? WHERE name=?", lastModified, eventKey)
	}
}
