package server

import (
	"encoding/json"
	"fmt"
	"io"
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

	eventsEndpoint := fmt.Sprintf("%s/events/%s/simple", tbaAPI, year)

	req, err := http.NewRequest("GET", eventsEndpoint, nil)
	if err != nil {
		return fmt.Errorf("error: TBA polling failed: %v", err)
	}

	lastModified, err := s.store.EventsModifiedData()
	if err == nil {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", apikey)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("error: TBA polling request failed: %v", err)
	}

	lastModified = response.Header.Get("Last-Modified")
	if lastModified != "" {
		if err := s.store.SetEventsModifiedData(lastModified); err != nil {
			return err
		}
	}

	responseCode := response.StatusCode

	if responseCode == http.StatusNotModified {
		logger.Debugf("TBA event data not modified")
		return nil
	} else if responseCode != http.StatusOK {
		return fmt.Errorf("TBA polling request failed with status %d", responseCode)
	}

	if err := json.NewDecoder(io.LimitReader(response.Body, 1.049e+6)).Decode(&tbaEvents); err != nil {
		return fmt.Errorf("error: reading TBA response: %v", err)
	}

	var events []store.Event

	for _, tbaEvent := range tbaEvents {
		date, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			logger.Debugf("error parsing TBA time data: %v\n", err)
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
		return fmt.Errorf("error: updating events: %v", err)
	}

	logger.Infof("Polled TBA...")
	return nil
}

func (s *Server) pollTBAMatches(tbaAPI string, apikey string, eventKey string) ([]store.Match, error) {
	type tbaMatch struct {
		Key string `json:"key"`
	}
	var tbaMatches []tbaMatch

	matchesEndpoint := fmt.Sprintf("%s/event/%s/matches/simple", tbaAPI, eventKey)

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
		if err := s.store.SetMatchModifiedData(lastModified, eventKey); err != nil {
			return nil, err
		}
	}

	responseCode := response.StatusCode

	if responseCode == http.StatusNotModified {
		return nil, nil
	} else if responseCode != http.StatusOK {
		return nil, fmt.Errorf("error: TBA polling request failed with status: %v", responseCode)
	}

	if err := json.NewDecoder(response.Body).Decode(&tbaMatches); err != nil {
		return nil, err
	}

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
