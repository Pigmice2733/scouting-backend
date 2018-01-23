package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Pigmice2733/scouting-backend/internal/respond"
	"github.com/Pigmice2733/scouting-backend/internal/server/logic"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type key int

const (
	keyUsernameCtx key = iota
	keyIsAdminCtx
)

func (s *Server) authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ss := strings.TrimPrefix(r.Header.Get("Authentication"), "Bearer ")
		token, err := jwt.Parse(ss, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return s.jwtSecret, nil
		})

		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var username string
		var isAdmin bool

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			var uOk, aOk bool
			username, uOk = claims[logic.SubjectClaim].(string)
			isAdmin, aOk = claims[logic.IsAdminClaim].(bool)

			if !uOk || !aOk {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), keyUsernameCtx, username)
		ctx = context.WithValue(ctx, keyIsAdminCtx, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type requestUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) authenticateHandler(w http.ResponseWriter, r *http.Request) {
	var reqUser requestUser

	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ss, err := logic.Authenticate(reqUser.Username, reqUser.Password, s.jwtSecret, s.store.User)
	if err != nil {
		if err == logic.ErrUnauthorized {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			s.logger.LogRequestError(r, fmt.Errorf("authenticating user: %v", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	respond.JSON(w, map[string]string{"jwt": ss})
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var reqUser requestUser

	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqUser.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("generating bcrypt hash from password: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := user.User{Username: reqUser.Username, HashedPassword: string(hashedPassword)}
	if err := s.store.User.Create(user); err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("creating user: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	if err := s.store.User.Delete(username); err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("deleting user: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
