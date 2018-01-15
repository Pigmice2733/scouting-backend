package logic

import (
	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/photo"
	"github.com/Pigmice2733/scouting-backend/internal/tba"
)

// GetPhoto gets a photo from either the photo store if it exists there, or,
// fetches it from the TBA API and attempts to store the photo in the photo
// store.
func GetPhoto(tbaKey, team string, year int, ps photo.Service) (string, error) {
	url, err := ps.Get(team)
	if err != nil && err != store.ErrNoResults {
		return "", err
	}

	if err == store.ErrNoResults || url == "" {
		url, err = tba.GetPhotoURL(tbaKey, team, year)
		if err != nil {
			return "", err
		}

		if url != "" {
			go ps.Create(team, url)
		}
	}

	return url, err
}
