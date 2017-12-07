package store

import (
	"fmt"

	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/match"
	"github.com/Pigmice2733/scouting-backend/internal/store/report"
	"github.com/Pigmice2733/scouting-backend/internal/store/tbamodified"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"
)

// ErrNoResults is a generic error of sql.ErrNoRows.
var ErrNoResults = fmt.Errorf("no results returned")

// Service provides an interface for interacting with a store.
type Service struct {
	Alliance    alliance.Service
	Event       event.Service
	Match       match.Service
	Report      report.Service
	TBAModified tbamodified.Service
	User        user.Service
}

// Close closes all services.
func (s *Service) Close() error {
	if err := s.Alliance.Close(); err != nil {
		return err
	}

	if err := s.Event.Close(); err != nil {
		return err
	}

	if err := s.Match.Close(); err != nil {
		return err
	}

	if err := s.Report.Close(); err != nil {
		return err
	}

	if err := s.TBAModified.Close(); err != nil {
		return err
	}

	return s.User.Close()
}
