package tba

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// ErrNotModified is returned if the tba data has not been modified since last retrieved.
var ErrNotModified = fmt.Errorf("tba data not modified")

// GetEvents retrieves all associated events from the blue alliance API.
func GetEvents(tbaURL, tbaKey, lastModified string, year int) ([]event.Event, string, error) {
	type tbaEvent struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Date string `json:"start_date"`
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/events/%s/simple", tbaURL, strconv.Itoa(year)), nil)
	if err != nil {
		return []event.Event{}, "", err
	}

	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", tbaKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []event.Event{}, "", err
	}

	if resp.StatusCode == http.StatusNotModified {
		return []event.Event{}, "", ErrNotModified
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []event.Event{}, "", fmt.Errorf("tba polling failed with status code: %d", resp.StatusCode)
	}

	var tbaEvents []tbaEvent
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1.049e+6)).Decode(&tbaEvents); err != nil {
		return []event.Event{}, "", err
	}

	var events []event.Event
	for _, tbaEvent := range tbaEvents {
		date, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			return events, "", nil
		}

		events = append(events, event.Event{
			Key:  tbaEvent.Key,
			Name: tbaEvent.Name,
			Date: date.UTC(),
		})
	}

	return events, resp.Header.Get("Last-Modified"), nil
}

// GetMatches retrieves all associated matches from the blue alliance API.
func GetMatches(tbaURL, tbaKey, eventKey, lastModified string) ([]match.Match, string, error) {
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

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/event/%s/matches/simple", tbaURL, eventKey), nil)
	if err != nil {
		return []match.Match{}, "", err
	}

	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", tbaKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []match.Match{}, "", err
	}

	if resp.StatusCode == http.StatusNotModified {
		return []match.Match{}, "", ErrNotModified
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []match.Match{}, "", fmt.Errorf("tba polling failed with status code: %d", resp.StatusCode)
	}

	var tbaMatches []tbaMatch
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1.049e+6)).Decode(&tbaMatches); err != nil {
		return []match.Match{}, "", err
	}

	var matches []match.Match
	for _, tbaMatch := range tbaMatches {
		var predictedMatchTime time.Time
		var actualMatchTime time.Time

		if tbaMatch.PredictedTime != 0 {
			predictedMatchTime = time.Unix(tbaMatch.PredictedTime, 0)
		} else if tbaMatch.ScheduledTime != 0 {
			predictedMatchTime = time.Unix(tbaMatch.ScheduledTime, 0)
		} else {
			predictedMatchTime = time.Time{}
		}

		if tbaMatch.ActualTime != 0 {
			actualMatchTime = time.Unix(tbaMatch.ActualTime, 0)
		} else {
			actualMatchTime = time.Time{}
		}

		matches = append(matches, match.Match{
			Key:             tbaMatch.Key,
			EventKey:        eventKey,
			PredictedTime:   predictedMatchTime.UTC(),
			ActualTime:      actualMatchTime.UTC(),
			WinningAlliance: tbaMatch.WinningAlliance,
			RedAlliance:     populateAlliance(false, tbaMatch.Key, tbaMatch.Alliances.Red),
			BlueAlliance:    populateAlliance(false, tbaMatch.Key, tbaMatch.Alliances.Blue),
		})
	}

	return matches, resp.Header.Get("Last-Modified"), nil
}

type allianceData struct {
	Score int      `json:"score"`
	Teams []string `json:"team_keys"`
}

func populateAlliance(isBlue bool, matchKey string, ad allianceData) alliance.Alliance {
	a := alliance.Alliance{
		MatchKey: matchKey,
		IsBlue:   isBlue,
		Score:    ad.Score,
	}

	for _, teamKey := range ad.Teams {
		team := alliance.Team{
			Number: teamKey,
		}
		a.Teams = append(a.Teams, team)
	}

	return a
}
