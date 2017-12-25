package store

import (
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
	"github.com/Pigmice2733/scouting-backend/internal/store/report"

	"github.com/Pigmice2733/scouting-backend/internal/store/match"

	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"
)

// ErrNoResults is a generic error of sql.ErrNoRows.
var ErrNoResults = fmt.Errorf("no results returned")

// Service provides an interface for interacting with a store.
type Service struct {
	Event    event.Service
	Match    match.Service
	Alliance alliance.Service
	Report   report.Service
	User     user.Service
}
