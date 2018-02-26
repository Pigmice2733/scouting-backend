package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
	"github.com/Pigmice2733/scouting-backend/internal/tba"
	"github.com/Pigmice2733/scouting-backend/internal/tba/api/lastmodified"
)

var lastModified = lastmodified.New()

const imgurFormat = "http://i.imgur.com/%sl.jpg"

// Consumer consumes info from the TBA api.
type Consumer struct {
	tbaURL string
	tbaKey string
}

// New returns a new TBA API Consumer.
func New(tbaURL, tbaKey string) *Consumer {
	return &Consumer{tbaURL: tbaURL, tbaKey: tbaKey}
}

func (c Consumer) makeRequest(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if lastModified := lastModified.Get(path); lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	req.Header.Set("X-TBA-Auth-Key", c.tbaKey)

	return http.DefaultClient.Do(req)
}

// GetEvents retrieves all associated events from the blue alliance API.
func (c Consumer) GetEvents(year int) ([]event.BasicEvent, error) {
	path := fmt.Sprintf("%s/events/%d", c.tbaURL, year)

	resp, err := c.makeRequest(path)
	if err != nil {
		return []event.BasicEvent{}, err
	}

	if resp.StatusCode == http.StatusNotModified {
		return []event.BasicEvent{}, tba.ErrNotModified
	} else if resp.StatusCode != http.StatusOK {
		return []event.BasicEvent{}, fmt.Errorf("tba: polling failed with status code: %d", resp.StatusCode)
	}

	var tbaEvents []tbaEvent
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1.049e+6)).Decode(&tbaEvents); err != nil {
		return []event.BasicEvent{}, err
	}

	var bEvents []event.BasicEvent
	for _, tbaEvent := range tbaEvents {
		timeZone, err := time.LoadLocation(tbaEvent.TimeZone)
		if err != nil {
			return bEvents, err
		}
		date, err := time.ParseInLocation("2006-01-02", tbaEvent.Date, timeZone)
		if err != nil {
			return bEvents, err
		}
		endDate, err := time.ParseInLocation("2006-01-02", tbaEvent.EndDate, timeZone)
		if err != nil {
			return bEvents, err
		}

		bEvents = append(bEvents, event.BasicEvent{
			Key:       tbaEvent.Key,
			Name:      tbaEvent.Name,
			ShortName: tbaEvent.ShortName,
			EventType: tbaEvent.EventType,
			Lat:       tbaEvent.Lat,
			Long:      tbaEvent.Lng,
			Date:      date,
			EndDate:   endDate,
		})
	}

	lastModified.Set(path, resp.Header.Get("Last-Modified"))

	return bEvents, nil
}

// GetMatches retrieves all associated matches from the blue alliance API.
func (c Consumer) GetMatches(eventKey string) ([]match.Match, error) {
	path := fmt.Sprintf("%s/event/%s/matches/simple", c.tbaURL, eventKey)

	resp, err := c.makeRequest(path)
	if err != nil {
		return []match.Match{}, err
	}

	if resp.StatusCode == http.StatusNotModified {
		return []match.Match{}, tba.ErrNotModified
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

func (c Consumer) getMedia(team string, year int) ([]media, error) {
	path := fmt.Sprintf("%s/team/%s/media/%d", c.tbaURL, team, year)

	resp, err := c.makeRequest(path)
	if err != nil {
		return []media{}, err
	}

	if resp.StatusCode == http.StatusNotModified {
		return []media{}, tba.ErrNotModified
	} else if resp.StatusCode == http.StatusNotFound {
		return []media{}, nil
	} else if resp.StatusCode != http.StatusOK {
		return []media{}, fmt.Errorf("tba: polling failed with status code: %d", resp.StatusCode)
	}

	var media []media
	err = json.NewDecoder(io.LimitReader(resp.Body, 1.049e+6)).Decode(&media)

	return media, err
}

// GetPhotoURL returns the optimal photo url for a team in a certain year from
// TBA api.
func (c Consumer) GetPhotoURL(team string, year int) (url string, err error) {
	media, err := c.getMedia(team, year)
	if err != nil {
		return "", err
	}

	for _, m := range media {
		if m.Type == "imgur" {
			url = fmt.Sprintf(imgurFormat, m.ForeignKey)
			break
		} else if m.Type == "instagram-image" {
			url = m.Details.ThumbnailURL
			break
		}
	}

	return
}
