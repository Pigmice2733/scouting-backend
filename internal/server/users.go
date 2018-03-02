package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Pigmice2733/scouting-backend/internal/store"

	"golang.org/x/crypto/bcrypt"

	"github.com/Pigmice2733/scouting-backend/internal/respond"
	"github.com/Pigmice2733/scouting-backend/internal/server/logic"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"
	"github.com/gorilla/mux"
)

type requestUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"isAdmin"`
}

type nullableRequestUser struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
	IsAdmin  *bool   `json:"isAdmin"`
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

func (s *Server) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.User.GetUsers()
	if err == store.ErrNoResults {
		users = []user.User{}
	} else if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		s.logger.LogRequestError(r, fmt.Errorf("getting all users: %v", err))
		return
	}

	var resp []map[string]interface{}
	for _, u := range users {
		resp = append(resp, map[string]interface{}{"username": u.Username, "isAdmin": u.IsAdmin})
	}

	respond.JSON(w, resp)
}

const allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var reqUser requestUser

	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	for _, ch := range reqUser.Username {
		if !strings.Contains(allowedChars, string(ch)) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqUser.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("generating bcrypt hash from password: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := user.User{Username: reqUser.Username, HashedPassword: string(hashedPassword), IsAdmin: reqUser.IsAdmin}
	if err := s.store.User.Create(user); err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("creating user: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	usernameToUpdate := mux.Vars(r)["username"]

	authenticatedUser, uOk := r.Context().Value(keyUsernameCtx).(string)
	authenticatedIsAdmin, aOk := r.Context().Value(keyIsAdminCtx).(bool)

	if !uOk || !aOk {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var reqUser nullableRequestUser
	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !authenticatedIsAdmin && (usernameToUpdate != authenticatedUser || (reqUser.IsAdmin != nil && *reqUser.IsAdmin)) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	updateUser := user.NullableUser{Username: reqUser.Username, IsAdmin: reqUser.IsAdmin}

	if reqUser.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*reqUser.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		hashededPasswordStr := string(hashedPassword)
		updateUser.HashedPassword = &hashededPasswordStr
	}

	if err := s.store.User.Update(usernameToUpdate, updateUser); err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("updating user: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	usernameToDelete := mux.Vars(r)["username"]

	authenticatedUser, uOk := r.Context().Value(keyUsernameCtx).(string)
	authenticatedIsAdmin, aOk := r.Context().Value(keyIsAdminCtx).(bool)

	if !uOk || !aOk {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !authenticatedIsAdmin && usernameToDelete != authenticatedUser {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err := s.store.User.Delete(usernameToDelete); err != nil {
		s.logger.LogRequestError(r, fmt.Errorf("deleting user: %v", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
