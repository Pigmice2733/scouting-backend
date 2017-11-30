package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store"
)

func (s *Server) pollTBAEvents(tbaAPI string, apikey string, year string) error {
	type tbaEvent struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Date string `json:"start_date"`
	}
	var tbaEvents []tbaEvent

	eventsEndpoint := fmt.Sprintf("%s/events/%s/simple", tbaAPI, year)

	req, err := http.NewRequest("GET", eventsEndpoint, nil)
	if err != nil {
		return fmt.Errorf("TBA polling failed: %v", err)
	}

	lastModified, err := s.store.EventsModifiedData()
	if err == nil {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", apikey)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("TBA polling request failed: %v", err)
	}

	lastModified = response.Header.Get("Last-Modified")
	if lastModified != "" {
		if err := s.store.SetEventsModifiedData(lastModified); err != nil {
			return err
		}
	}

	responseCode := response.StatusCode

	if responseCode == http.StatusNotModified {
		s.logger.Debugf("TBA event data not modified")
		return nil
	} else if responseCode != http.StatusOK {
		return fmt.Errorf("TBA polling request failed with status %d", responseCode)
	}

	if err := json.NewDecoder(io.LimitReader(response.Body, 1.049e+6)).Decode(&tbaEvents); err != nil {
		return fmt.Errorf("reading TBA response: %v", err)
	}

	var events []store.Event

	for _, tbaEvent := range tbaEvents {
		date, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			s.logger.Debugf("error parsing TBA time data: %v\n", err)
			continue
		}
		newEvent := store.Event{
			Key:  tbaEvent.Key,
			Name: tbaEvent.Name,
			Date: date.UTC(),
		}
		events = append(events, newEvent)
	}

	errs := s.store.UpdateEvents(events)
	if len(errs) != 0 {
		for _, err := range errs {
			s.logger.Errorf("error: updating events: %v\n", err)
		}
		return fmt.Errorf("error: updating events")
	}

	s.logger.Infof("Polled TBA...")
	return nil
}

func (s *Server) pollTBAMatches(tbaAPI string, apikey string, eventKey string) error {
	type tbaMatch struct {
		Key             string `json:"key"`
		ScheduledTime   int64  `json:"time"`
		PredictedTime   int64  `json:"predicted_time"`
		ActualTime      int64  `json:"actual_time"`
		WinningAlliance string `json:"winning_alliance"`
		Alliances       struct {
			Blue struct {
				Score int      `json:"score"`
				Teams []string `json:"team_keys"`
			} `json:"blue"`
			Red struct {
				Score int      `json:"score"`
				Teams []string `json:"team_keys"`
			} `json:"red"`
		} `json:"alliances"`
	}
	var tbaMatches []tbaMatch

	matchesEndpoint := fmt.Sprintf("%s/event/%s/matches/simple", tbaAPI, eventKey)

	req, err := http.NewRequest("GET", matchesEndpoint, nil)
	if err != nil {
		return err
	}

	lastModified, err := s.store.MatchModifiedData(eventKey)
	if err == nil {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", apikey)

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return err
	}

	lastModified = response.Header.Get("Last-Modified")
	if lastModified != "" {
		if err := s.store.SetMatchModifiedData(lastModified, eventKey); err != nil {
			return err
		}
	}

	responseCode := response.StatusCode

	if responseCode == http.StatusNotModified {
		return nil
	} else if responseCode != http.StatusOK {
		return fmt.Errorf("error: TBA polling request failed with status: %v", responseCode)
	}

	if err := json.NewDecoder(response.Body).Decode(&tbaMatches); err != nil {
		return err
	}

	var matches []store.Match

	for _, tbaMatch := range tbaMatches {
		var predictedMatchTime time.Time
		var actualMatchTime time.Time

		if !time.Unix(tbaMatch.PredictedTime, 0).IsZero() {
			predictedMatchTime = time.Unix(tbaMatch.PredictedTime, 0)
		} else if !time.Unix(tbaMatch.ScheduledTime, 0).IsZero() {
			predictedMatchTime = time.Unix(tbaMatch.ScheduledTime, 0)
		} else {
			predictedMatchTime = time.Time{}
		}

		if !time.Unix(tbaMatch.ActualTime, 0).IsZero() {
			actualMatchTime = time.Unix(tbaMatch.ActualTime, 0)
		} else {
			actualMatchTime = time.Time{}
		}

		newMatch := store.Match{
			Key:             tbaMatch.Key,
			EventKey:        eventKey,
			PredictedTime:   predictedMatchTime.UTC(),
			ActualTime:      actualMatchTime.UTC(),
			WinningAlliance: tbaMatch.WinningAlliance,
		}

		newMatch.RedAlliance, err = populateAlliance(false, tbaMatch.Key, tbaMatch.Alliances.Red)
		if err != nil {
			return err
		}
		newMatch.BlueAlliance, err = populateAlliance(true, tbaMatch.Key, tbaMatch.Alliances.Blue)
		if err != nil {
			return err
		}

		matches = append(matches, newMatch)
	}

	errs := s.store.UpdateMatches(matches)
	if len(errs) != 0 {
		for _, err := range errs {
			s.logger.Errorf("error: updating events: %v\n", err)
		}
		return fmt.Errorf("error: updating events")
	}

	return nil
}

// Use data from tba to populate an alliance struct
func populateAlliance(isBlue bool, matchKey string, allianceData struct {
	Score int      `json:"score"`
	Teams []string `json:"team_keys"`
}) (store.Alliance, error) {
	alliance := store.Alliance{
		MatchKey: matchKey,
		IsBlue:   isBlue,
		Score:    allianceData.Score,
	}

	for _, teamKey := range allianceData.Teams {
		team := store.TeamInAlliance{
			Number: teamKey,
		}

		alliance.Teams = append(alliance.Teams, team)
	}

	return alliance, nil
}
