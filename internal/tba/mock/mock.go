package mock

import (
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// ErrNoSuchYear is returned when a year does not exist in the mock db.
var ErrNoSuchYear = fmt.Errorf("mock: no such year")

// ErrNoSuchEvent is returned when an event does not exist in the mock db.
var ErrNoSuchEvent = fmt.Errorf("mock: no such event")

// ErrNoSuchTeam is returned when a team does not exist in the mock db.
var ErrNoSuchTeam = fmt.Errorf("mock: no such team")

// DB mocks a TBA API Consumer.
type DB struct {
	Events  map[int][]event.BasicEvent // year --> events
	Matches map[string][]match.Match   // eventkey --> matches
	Photos  map[int]map[string]string  // year --> eventkey --> photo
}

// GetEvents gets all events in the mock db for TBA by year.
func (db DB) GetEvents(year int) ([]event.BasicEvent, error) {
	if events, ok := db.Events[year]; ok {
		return events, nil
	}
	return []event.BasicEvent{}, ErrNoSuchYear
}

// GetMatches gets all matches in the mock db for TBA by eventKey.
func (db DB) GetMatches(eventKey string) ([]match.Match, error) {
	if matches, ok := db.Matches[eventKey]; ok {
		return matches, nil
	}
	return []match.Match{}, ErrNoSuchEvent
}

// GetPhotoURL gets the photo URL in the mock db for TBA by year and team.
func (db DB) GetPhotoURL(team string, year int) (string, error) {
	if teams, ok := db.Photos[year]; ok {
		if url, ok := teams[team]; ok {
			return url, nil
		}
		return "", ErrNoSuchTeam
	}
	return "", ErrNoSuchYear
}
