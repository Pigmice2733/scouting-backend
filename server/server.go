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

	"golang.org/x/crypto/bcrypt"

	"context"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/server/logger"
	"github.com/Pigmice2733/scouting-backend/server/store"
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
	Handler   http.Handler
	store     store.Service
	logger    logger.Service
	tbaAPIKey string
	jwtSecret []byte
}

// New creates a new server given a db file and a io.Writer for logging
func New(store store.Service, logWriter io.Writer, tbaAPIKey string, environment string) (*Server, error) {
	s := &Server{}

	s.logger = logger.New(logWriter, logger.Settings{
		Debug: environment == "dev",
		Info:  true,
		Error: true,
	})

	s.store = store

	s.initializeRouter()
	s.initializeMiddlewares()

	s.tbaAPIKey = tbaAPIKey

	var err error
	s.jwtSecret, err = generateSecret(64)
	if err != nil {
		return s, fmt.Errorf("error: generating jwt secret: %v", err)
	}

	return s, nil
}

// Run starts a running the server on the specified address
func (s *Server) Run(addr string) error {
	s.logger.Infof("server up and running")
	return http.ListenAndServe(addr, s.Handler)
}

// PollTBA polls The Blue Alliance api to populate database
func (s *Server) PollTBA(year string) error {
	return s.pollTBAEvents(s.logger, "https://www.thebluealliance.com/api/v3", s.tbaAPIKey, year)
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

	s.logger.Infof("initialized router...")
}

func (s *Server) initializeMiddlewares() {
	s.Handler = tollbooth.LimitHandler(tollbooth.NewLimiter(1, nil),
		(limitHandler(
			s.logger.Middleware(
				gziphandler.GzipHandler(
					addDefaultHeaders(s.Handler))))))
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

	user, err := s.store.GetUser(authInfo.Username)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			s.logger.Errorf("error: finding user in database: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(authInfo.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			s.logger.Errorf("error: comparing password hashes: %v\n", err)
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

	user, err := s.store.GetUser(username)
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
	users, err := s.store.GetUsers()
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting users: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(users)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1440)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.Errorf("error: handling eTag %v", err)
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

	user := store.User{Username: authInfo.Username, HashedPassword: string(hashedPassword)}
	if err := s.store.CreateUser(user); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	noCache(w, 0)

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	if err := s.store.DeleteUser(username); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	noCache(w, 0)
}

func (s *Server) getEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.store.GetEvents()
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting events: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(events)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1440)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.Errorf("error: handling eTag %v", err)
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

	e, err := s.store.GetEvent(eventKey)
	if err != nil {
		if err == store.ErrNoResults {
			http.Error(w, "non-existent event key", http.StatusNotFound)
		} else {
			s.logger.Errorf("error: getting events: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	err = s.pollTBAMatches("https://www.thebluealliance.com/api/v3", s.tbaAPIKey, e.Key)
	if err != nil {
		s.logger.Infof(err.Error())
	}

	matches, err := s.store.GetAllMatchData(e.Key)
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting events: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fullEvent := store.FullEvent{
		Key:     e.Key,
		Name:    e.Name,
		Date:    e.Date,
		Matches: matches,
	}

	response, err := json.Marshal(fullEvent)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.Errorf("error: handling eTag %v", err)
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

	match, err := s.store.GetMatch(eventKey, matchKey)

	if err != nil {
		if err == store.ErrNoResults {
			message := fmt.Sprintf("non-existent match key '%v' under event key '%v'", matchKey, eventKey)
			http.Error(w, message, http.StatusNotFound)
		} else {
			s.logger.Errorf("error: getting match: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	blueAlliance, _, err := s.store.GetAlliance(matchKey, true)
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting alliance: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	redAlliance, _, err := s.store.GetAlliance(matchKey, false)
	if err != nil && err != store.ErrNoResults {
		s.logger.Errorf("error: getting alliance: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	blueTeams, err := s.store.GetTeamsInAlliance(blueAlliance.ID)
	if err != nil {
		s.logger.Errorf("error: getting alliance teams: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	redTeams, err := s.store.GetTeamsInAlliance(redAlliance.ID)
	if err != nil {
		s.logger.Errorf("error: getting alliance teams: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	blueAlliance.Teams = blueTeams
	redAlliance.Teams = redTeams

	match.BlueAlliance = blueAlliance
	match.RedAlliance = redAlliance

	response, err := json.Marshal(match)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	foundMatchingEtag, err := addETags(w, r, response)
	if err != nil {
		s.logger.Errorf("error: handling eTag %v", err)
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

	var report store.ReportData

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

	matchExists, err := s.store.CheckMatchExistence(eventKey, matchKey)
	if err != nil {
		s.logger.Errorf("error: check if match exists %v. matchKey '%v' eventKey '%v'\n", err, matchKey, eventKey)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if !matchExists {
		http.Error(w, "error: posting report to non-existent match", http.StatusBadRequest)
		return
	}

	alllianceID, a, err := s.findAlliance(matchKey, report)
	if err != nil {
		s.logger.Errorf("ERROR: Alliance for match somehow not created! %v %v", matchKey, report)
		if err != store.ErrNoResults {
			s.logger.Errorf("error: nothing present in response: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		team := store.TeamInAlliance{
			Number: report.Team,
		}
		a.Teams = []store.TeamInAlliance{team}
		allianceID, err := s.store.CreateAlliance(a)
		a.ID = allianceID
		if err != nil {
			http.Error(w, "no corresponding match for posted report", http.StatusBadRequest)
			return
		}
		err = s.store.CreateTeamInAlliance(allianceID, team)
		if err != nil {
			http.Error(w, "failed to create team in alliance", http.StatusBadRequest)
			return
		}
	} else {
		a.ID = alllianceID
		teams, err := s.store.GetTeamsInAlliance(a.ID)
		for _, team := range teams {
			if report.Team == team.Number {
				http.Error(w, "report for team already exists, use 'PUT' to update", http.StatusConflict)
				return
			}
		}

		a.Score = report.Score
		err = s.store.UpdateAlliance(a)
		if err != nil {
			http.Error(w, "failed to update alliance data", http.StatusBadRequest)
			return
		}

		team := store.TeamInAlliance{
			Number: report.Team,
		}
		a.Teams = []store.TeamInAlliance{team}
		err = s.store.CreateTeamInAlliance(a.ID, team)
		if err != nil {
			http.Error(w, "failed to create team in alliance", http.StatusBadRequest)
			return
		}
	}

	if err := s.store.CreateReport(report, a.ID); err != nil {
		s.logger.Errorf("error: postreport: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(report)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
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

	var report store.ReportData

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
		s.logger.Errorf("error: updateReport %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	a.ID = allianceID
	teams, err := s.store.GetTeamsInAlliance(a.ID)
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

	if err := s.store.UpdateReport(report, a.ID); err != nil {
		s.logger.Errorf("error: updateReport %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(report)
	if err != nil {
		s.logger.Errorf("error: marshalling json response %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	normalCache(w, 1)

	if _, err := w.Write(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) findAlliance(matchKey string, report store.ReportData) (int, store.Alliance, error) {
	isBlue := true
	if report.Alliance == "red" {
		isBlue = false
	}

	a, id, err := s.store.GetAlliance(matchKey, isBlue)

	return id, a, err
}

func (s *Server) createEvents(events []store.Event) error {
	for _, event := range events {
		err := s.store.CreateEvent(event)
		if err != nil {
			s.logger.Errorf("error: processing TBA data '%v' in data '%v'\n", err, event)
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
		s.logger.Errorf("error: signing jwt string: %v\n", err)
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
