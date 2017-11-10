package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Pigmice2733/scouting-backend/logger"
	"github.com/Pigmice2733/scouting-backend/store"
)

func (s *Server) pollTBAEvents(logger logger.Service, tbaAPI string, apikey string, year string) error {
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

	lastModified, err := s.store.EventsModifiedData()
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
		s.store.SetEventsModifiedData(lastModified)
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

	var events []store.Event

	for _, tbaEvent := range tbaEvents {
		date, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			logger.Debugf("Error TBA time data %v", err.Error())
			continue
		}
		newEvent := store.Event{
			Key:  tbaEvent.Key,
			Name: tbaEvent.Name,
			Date: date,
		}
		events = append(events, newEvent)
	}

	err = s.store.UpdateEvents(events)
	if err != nil {
		return err
	}

	logger.Infof("Polled TBA...")
	return nil
}

func (s *Server) pollTBAMatches(tbaAPI string, apikey string, eventKey string) ([]store.Match, error) {
	type tbaMatch struct {
		Key string `json:"key"`
	}
	var tbaMatches []tbaMatch

	matchesEndpoint := fmt.Sprintf("%s/event/%s/matches", tbaAPI, eventKey)

	req, err := http.NewRequest("GET", matchesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	lastModified, err := s.store.MatchModifiedData(eventKey)
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
		s.store.SetMatchModifiedData(lastModified, eventKey)
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

	var matches []store.Match

	for _, tbaMatch := range tbaMatches {
		newMatch := store.Match{
			Key:      tbaMatch.Key,
			EventKey: eventKey,
		}
		matches = append(matches, newMatch)
	}

	err = s.store.UpdateMatches(matches)
	if err != nil {
		return nil, err
	}

	return matches, nil
}
