package tba

import (
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
)

// ErrNotModified is returned if the tba data has not been modified since last retrieved.
var ErrNotModified = fmt.Errorf("tba data not modified")

// Consumer provides an interface for getting information from TBA api.
type Consumer interface {
	GetEvents(year int) ([]event.BasicEvent, error)
	GetMatches(eventKey string) ([]match.Match, error)
	GetPhotoURL(team string, year int) (url string, err error)
}
