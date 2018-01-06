package tba

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// ErrNotModified is returned if the tba data has not been modified since last retrieved.
var ErrNotModified = fmt.Errorf("tba data not modified")

type lastModifiedManager struct {
	pathLastModified map[string]string
	mu               *sync.Mutex
}

func (lmm lastModifiedManager) Get(path string) (lastModified string) {
	lmm.mu.Lock()
	lastModified = lmm.pathLastModified[path]
	lmm.mu.Unlock()
	return
}

func (lmm *lastModifiedManager) Set(path, lastModified string) {
	lmm.mu.Lock()
	lmm.pathLastModified[path] = lastModified
	lmm.mu.Unlock()
}

var lastModified = lastModifiedManager{make(map[string]string), new(sync.Mutex)}

const tbaURL = "http://www.thebluealliance.com/api/v3"

// GetEvents retrieves all associated events from the blue alliance API.
func GetEvents(tbaKey string, year int) ([]event.BasicEvent, error) {
	type tbaEvent struct {
		Key       string `json:"key"`
		Name      string `json:"name"`
		ShortName string `json:"short_name"`
		EventType int    `json:"event_type"`
		Date      string `json:"start_date"`
	}

	path := fmt.Sprintf("%s/events/%d", tbaURL, year)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return []event.BasicEvent{}, err
	}

	if lastModified := lastModified.Get(path); lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", tbaKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []event.BasicEvent{}, err
	}

	if resp.StatusCode == http.StatusNotModified {
		return []event.BasicEvent{}, ErrNotModified
	} else if resp.StatusCode != http.StatusOK {
		return []event.BasicEvent{}, fmt.Errorf("tba: polling failed with status code: %d", resp.StatusCode)
	}

	var tbaEvents []tbaEvent
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1.049e+6)).Decode(&tbaEvents); err != nil {
		return []event.BasicEvent{}, err
	}

	var bEvents []event.BasicEvent
	for _, tbaEvent := range tbaEvents {
		startDate, err := time.Parse("2006-01-02", tbaEvent.Date)
		if err != nil {
			return bEvents, err
		}

		bEvents = append(bEvents, event.BasicEvent{
			Key:       tbaEvent.Key,
			Name:      tbaEvent.Name,
			ShortName: tbaEvent.ShortName,
			EventType: tbaEvent.EventType,
			Date:      startDate,
		})
	}

	lastModified.Set(path, resp.Header.Get("Last-Modified"))

	return bEvents, nil
}

// GetMatches retrieves all associated matches from the blue alliance API.
func GetMatches(tbaKey, eventKey string) ([]match.Match, error) {
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

	path := fmt.Sprintf("%s/event/%s/matches/simple", tbaURL, eventKey)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return []match.Match{}, err
	}

	if lastModified := lastModified.Get(path); lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", tbaKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []match.Match{}, err
	}

	if resp.StatusCode == http.StatusNotModified {
		return []match.Match{}, ErrNotModified
	} else if resp.StatusCode != http.StatusOK {
		return []match.Match{}, fmt.Errorf("tba: polling failed with status code: %d", resp.StatusCode)
	}

	var tbaMatches []tbaMatch
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1.049e+6)).Decode(&tbaMatches); err != nil {
		return []match.Match{}, err
	}

	var bMatches []match.Match
	for _, tbaMatch := range tbaMatches {
		var predictedMatchTime *time.Time
		var actualMatchTime *time.Time

		if tbaMatch.PredictedTime != 0 {
			predictedTime := time.Unix(tbaMatch.PredictedTime, 0)
			predictedMatchTime = &predictedTime
		} else if tbaMatch.ScheduledTime != 0 {
			scheduledTime := time.Unix(tbaMatch.ScheduledTime, 0)
			predictedMatchTime = &scheduledTime
		}

		if tbaMatch.ActualTime != 0 {
			actualTime := time.Unix(tbaMatch.ActualTime, 0)
			actualMatchTime = &actualTime
		}

		bMatches = append(bMatches, match.Match{
			BasicMatch: match.BasicMatch{
				Key:           tbaMatch.Key,
				EventKey:      eventKey,
				PredictedTime: predictedMatchTime,
				ActualTime:    actualMatchTime,
			},
			RedScore:     tbaMatch.Alliances.Red.Score,
			BlueScore:    tbaMatch.Alliances.Blue.Score,
			RedAlliance:  tbaMatch.Alliances.Red.Teams,
			BlueAlliance: tbaMatch.Alliances.Blue.Teams,
		})
	}

	lastModified.Set(path, resp.Header.Get("Last-Modified"))

	return bMatches, nil
}
