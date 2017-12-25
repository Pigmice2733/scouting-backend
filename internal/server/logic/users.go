package logic

import (
	"fmt"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// ErrUnauthorized is returned when the user being attempted to be authorized has either the wrong username or password.
var ErrUnauthorized = fmt.Errorf("unauthorized user")

// Authenticate gives a jwt signed string for a certain user.
func Authenticate(username, password string, jwtSecret []byte, us user.Service) (string, error) {
	user, err := us.Get(username)
	if err != nil {
		if err == store.ErrNoResults {
			return "", ErrUnauthorized
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return "", ErrUnauthorized
		}
		return "", err
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   username,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}).SignedString(jwtSecret)
}
