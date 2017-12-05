package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/tba"

	"github.com/Pigmice2733/scouting-backend/internal/store/alliance"
	"github.com/Pigmice2733/scouting-backend/internal/store/event"
	"github.com/Pigmice2733/scouting-backend/internal/store/report"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"

	"golang.org/x/crypto/bcrypt"

	"context"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/internal/logger"
	"github.com/Pigmice2733/scouting-backend/internal/store"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/sha3"
)

type key int

const (
	keyUsernameCtx key = iota
)

// A Server is an instance of the scouting server
type Server struct {
	Handler     http.Handler
	store       *store.Service
	logger      logger.Service
	maxHandlers int
	tbaAPIKey   string
	jwtSecret   []byte
}

// New creates a new server given a db file and a io.Writer for logging
func New(store *store.Service, logWriter io.Writer, tbaAPIKey string, maxHandlers int) (*Server, error) {
	s := &Server{logger: logger.New(logWriter), store: store, tbaAPIKey: tbaAPIKey, maxHandlers: maxHandlers}

	s.initializeRouter()
	s.initializeMiddlewares()

	jwtSecret, err := generateSecret(64)
	if err != nil {
		return s, fmt.Errorf("generating jwt secret: %v", err)
	}
	s.jwtSecret = jwtSecret

	return s, nil
}

// Run starts a running the server on the specified address
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Handler)
}

func (s *Server) initializeRouter() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/authenticate", s.authenticate).Methods("POST")
	router.Handle("/users", s.authHandler(http.HandlerFunc(s.getUsers))).Methods("GET")
	router.Handle("/users", s.authHandler(http.HandlerFunc(s.createUser))).Methods("POST")
	router.Handle("/users/{username}", s.authHandler(http.HandlerFunc(s.getUser))).Methods("GET")
	router.Handle("/users/{username}", s.authHandler(http.HandlerFunc(s.deleteUser))).Methods("DELETE")
	router.HandleFunc("/events", s.getEvents).Methods("GET")
	router.HandleFunc("/events/{eventKey}", s.getEvent).Methods("GET")
	router.HandleFunc("/events/{eventKey}/{matchKey}", s.getMatch).Methods("GET")
	router.Handle("/events/{eventKey}/{matchKey}", s.authHandler(http.HandlerFunc(s.postReport))).Methods("POST")
	router.Handle("/events/{eventKey}/{matchKey}/{team}", s.authHandler(http.HandlerFunc(s.updateReport))).Methods("PUT")

	s.Handler = router
}

func (s *Server) initializeMiddlewares() {
	s.Handler = tollbooth.LimitHandler(tollbooth.NewLimiter(1, nil),
		limitHandler(
			gziphandler.GzipHandler(
				addDefaultHeaders(s.Handler))))
}

// REST Endpoint Handlers -----------------------

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	authInfo := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&authInfo); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := s.store.User.Get(authInfo.Username)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("retrieving user: %v", err)})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(authInfo.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("comparing password hashes: %v", err)})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// from this point on the user has been successfully authenticated, just give them a token!

	ss, err := s.GenerateJWT(user.Username)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	noCache(w, 0)

	json.NewEncoder(w).Encode(map[string]string{"jwt": ss})
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	user, err := s.store.User.Get(username)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	normalCache(w, 1440)

	json.NewEncoder(w).Encode(user)
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.User.GetUsers()
	if err != nil && err != store.ErrNoResults {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting users from database: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(users)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("marshalling json response: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1440)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("adding eTag: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if foundMatchingEtag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if _, err := w.Write(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	authInfo := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&authInfo); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(authInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user := user.User{Username: authInfo.Username, HashedPassword: string(hashedPassword)}
	if err := s.store.User.Create(user); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	noCache(w, 0)

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	if err := s.store.User.Delete(username); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	noCache(w, 0)
}

func (s *Server) getEvents(w http.ResponseWriter, r *http.Request) {
	lastModified, err := s.store.TBAModified.EventsModified()
	if err == store.ErrNoResults {
		lastModified = ""
	} else if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": err})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	events, newModified, err := tba.GetEvents("https://www.thebluealliance.com/api/v3", s.tbaAPIKey, lastModified, time.Now().Year())
	if err == nil {
		if err := s.store.TBAModified.SetEventsModified(newModified); err != nil {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("setting when events were last modified: %v", err)})
		}
		if errs := s.store.Event.UpdateEvents(events, s.maxHandlers); len(errs) != 0 {
			s.logger.LogRequestJSON(r, map[string]interface{}{"errors": errs})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else if err == tba.ErrNotModified {
		events, err = s.store.Event.GetEvents()
		if err != nil && err != store.ErrNoResults {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting events: %v", err)})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("polling tba: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(events)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("marshalling json response: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1440)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("adding eTag: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if foundMatchingEtag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if _, err := w.Write(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	eventKey := vars["eventKey"]

	e, err := s.store.Event.Get(eventKey)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, "non-existent event key", http.StatusNotFound)
		} else {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting events: %v", err)})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	lastModified, err := s.store.TBAModified.MatchModified(eventKey)
	if err != nil {
		if err == store.ErrNoResults {
			lastModified = ""
		} else {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting when matches were last modified: %v", err)})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	tbaMatches, newModified, err := tba.GetMatches("https://www.thebluealliance.com/api/v3", s.tbaAPIKey, e.Key, lastModified)
	if err != nil && err != tba.ErrNotModified {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("polling tba: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if err == nil {
		if err := s.store.TBAModified.SetMatchModified(eventKey, newModified); err != nil {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("setting when matches were last modified: %v", err)})
		}
		if errs := s.store.Match.UpdateMatches(tbaMatches, s.maxHandlers, s.store.Alliance); len(errs) != 0 {
			s.logger.LogRequestJSON(r, map[string]interface{}{"errors": err})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	matches, err := s.store.Match.GetMatches(e.Key, s.store.Alliance)
	if err != nil && err != store.ErrNoResults {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting matches: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fullEvent := event.Event{
		Key:     e.Key,
		Name:    e.Name,
		Date:    e.Date,
		Matches: matches,
	}

	response, err := json.Marshal(fullEvent)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("marshalling json: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("adding eTag: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if foundMatchingEtag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if _, err := w.Write(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) getMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	eventKey := vars["eventKey"]
	matchKey := vars["matchKey"]

	match, err := s.store.Match.Get(eventKey, matchKey)

	if err != nil {
		if err == store.ErrNoResults {
			message := fmt.Sprintf("non-existent match key '%v' under event key '%v'", matchKey, eventKey)
			http.Error(w, message, http.StatusNotFound)
		} else {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting match: %v", err)})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	blueAlliance, err := s.store.Alliance.Get(matchKey, true)
	if err != nil && err != store.ErrNoResults {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting alliance: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	redAlliance, err := s.store.Alliance.Get(matchKey, false)
	if err != nil && err != store.ErrNoResults {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting alliance: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	blueTeams, err := s.store.Alliance.GetTeams(blueAlliance.ID)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting alliance teams: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	redTeams, err := s.store.Alliance.GetTeams(redAlliance.ID)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("getting alliance teams: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	blueAlliance.Teams = blueTeams
	redAlliance.Teams = redTeams

	match.BlueAlliance = blueAlliance
	match.RedAlliance = redAlliance

	response, err := json.Marshal(match)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("marshalling json: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("adding eTag: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if foundMatchingEtag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if _, err := w.Write(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) postReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	matchKey := vars["matchKey"]
	eventKey := vars["eventKey"]

	var report report.Report

	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if reporter, ok := r.Context().Value(keyUsernameCtx).(string); ok {
		report.Reporter = reporter
	} else {
		report.Reporter = ""
	}

	matchExists, err := s.store.Match.Exists(eventKey, matchKey)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("checking if match exists: %v", err), "matchKey": matchKey, "eventKey": eventKey})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if !matchExists {
		http.Error(w, "error: posting report to non-existent match", http.StatusBadRequest)
		return
	}

	alllianceID, a, err := s.findAlliance(matchKey, report)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("alliance for match not created: %v", err)})
		if err != store.ErrNoResults {
			s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("nothing present in response: %v", err)})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		team := alliance.Team{Number: report.Team}
		a.Teams = []alliance.Team{team}
		allianceID, err := s.store.Alliance.Create(a)
		a.ID = allianceID
		if err != nil {
			http.Error(w, "no corresponding match for posted report", http.StatusBadRequest)
			return
		}
		err = s.store.Alliance.CreateTeam(allianceID, team)
		if err != nil {
			http.Error(w, "failed to create team in alliance", http.StatusBadRequest)
			return
		}
	} else {
		a.ID = alllianceID
		teams, err := s.store.Alliance.GetTeams(a.ID)
		for _, team := range teams {
			if report.Team == team.Number {
				http.Error(w, "report for team already exists, use 'PUT' to update", http.StatusConflict)
				return
			}
		}

		a.Score = report.Score
		err = s.store.Alliance.Update(a)
		if err != nil {
			http.Error(w, "failed to update alliance data", http.StatusBadRequest)
			return
		}

		team := alliance.Team{Number: report.Team}

		a.Teams = []alliance.Team{team}
		err = s.store.Alliance.CreateTeam(a.ID, team)
		if err != nil {
			http.Error(w, "failed to create team in alliance", http.StatusBadRequest)
			return
		}
	}

	if err := s.store.Report.Create(report, a.ID); err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("creating report: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(report)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("marshalling response: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) updateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	matchKey := vars["matchKey"]
	teamNumber := vars["team"]

	var report report.Report

	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if reporter, ok := r.Context().Value(keyUsernameCtx).(string); ok {
		report.Reporter = reporter
	} else {
		report.Reporter = ""
	}

	allianceID, a, err := s.findAlliance(matchKey, report)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, "report does not exist, use 'POST' to create a new report", http.StatusNotFound)
			return
		}
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("updating report: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	a.ID = allianceID
	teams, err := s.store.Alliance.GetTeams(a.ID)
	teamReportExists := false
	for _, team := range teams {
		if report.Team == team.Number {
			teamReportExists = true
			break
		}
	}
	if !teamReportExists {
		http.Error(w, "report does not exist, use 'POST' to create a new report", http.StatusNotFound)
		return
	}

	if report.Team != teamNumber {
		http.Error(w, "report team does not match team specified in URI", http.StatusBadRequest)
		return
	}

	if err := s.store.Report.Update(report, a.ID); err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("updating report: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(report)
	if err != nil {
		s.logger.LogRequestJSON(r, map[string]interface{}{"error": fmt.Sprintf("marshalling json: %v", err)})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	if _, err := w.Write(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) findAlliance(matchKey string, report report.Report) (int, alliance.Alliance, error) {
	isBlue := true
	if report.Alliance == "red" {
		isBlue = false
	}

	a, err := s.store.Alliance.Get(matchKey, isBlue)

	return a.ID, a, err
}

func (s *Server) createEvents(events []event.Event) error {
	for _, event := range events {
		err := s.store.Event.Create(event)
		if err != nil {
			s.logger.LogJSON(map[string]interface{}{"event": event, "error": fmt.Sprintf("creating event: %v", err)})
			return err
		}
	}
	return nil
}

func normalCache(w http.ResponseWriter, cacheMinutes int) {
	cacheControl := fmt.Sprintf("public, max-age=%d", (cacheMinutes * 60))
	if cacheMinutes == 0 {
		cacheControl = "no-cache"
	}

	w.Header().Set("Cache-Control", cacheControl)
}

func noCache(w http.ResponseWriter, cacheMinutes int) {
	w.Header().Set("Cache-Control", "no-store")
}

func addETags(w http.ResponseWriter, r *http.Request, response []byte) (bool, error) {
	contentETag, err := generateEtag(response)
	if err != nil {
		return false, err
	}

	ifNoneMatch := r.Header["If-None-Match"]

	for _, eTag := range ifNoneMatch {
		if eTag == contentETag {
			return true, nil
		}
	}

	w.Header().Set("ETag", contentETag)

	return false, nil
}

func addDefaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

// GenerateJWT creates a new token for an authenticated user
func (s *Server) GenerateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   username,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	ss, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.LogJSON(map[string]interface{}{"user": username, "error": fmt.Sprintf("signing jwt string: %v", err)})
		return "", err
	}

	return ss, nil
}

func limitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		next.ServeHTTP(w, r)
	})
}

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
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if username, ok = claims["sub"].(string); !ok {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), keyUsernameCtx, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateEtag(content []byte) (string, error) {
	hash := make([]byte, 32)
	d := sha3.NewShake256()
	// Write the response into the hash
	if _, err := d.Write(content); err != nil {
		return "", err
	}
	// Read 32 bytes of output from the hash into h.
	if _, err := d.Read(hash); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash), nil
}

func isError(statusCode int) bool {
	// HTTP error codes are status codes 4xx-5xx
	if statusCode >= 400 && statusCode < 600 {
		return true
	}
	return false
}

func generateSecret(secretLength int) ([]byte, error) {
	secret := make([]byte, 64)
	_, err := rand.Read(secret)
	return secret, err
}
