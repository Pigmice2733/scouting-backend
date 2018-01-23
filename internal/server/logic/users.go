package logic

import (
	"fmt"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	// SubjectClaim specifies the subject (user) of the jwt
	SubjectClaim = "sub"
	// ExpiresAtClaim specifies the time at which the jwt will expire
	ExpiresAtClaim = "exp"
	// IsAdminClaim specifies if the user is an admin user, and is prefixed with
	// the pigmice prefix for collision resistance
	IsAdminClaim = "pigmice_is_admin"
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

	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		SubjectClaim:   username,
		ExpiresAtClaim: time.Now().Add(time.Hour * 24).Unix(),
		IsAdminClaim:   user.IsAdmin,
	}).SignedString(jwtSecret)
}
