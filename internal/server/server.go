package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/tba"

	"github.com/Pigmice2733/scouting-backend/internal/analysis"
	"github.com/Pigmice2733/scouting-backend/internal/logger"
	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/gorilla/mux"
)

// A Server is an instance of the scouting server
type Server struct {
	handler   http.Handler
	store     *store.Service
	logger    logger.Service
	schema    analysis.Schema
	tbaAPIKey string
	jwtSecret []byte
	certFile  string
	keyFile   string
}

// New creates a new server given a db file and a io.Writer for logging
func New(store *store.Service, logWriter io.Writer, tbaAPIKey, schemaPath string, certFile, keyFile string) (*Server, error) {
	s := &Server{
		logger:    logger.New(logWriter),
		store:     store,
		tbaAPIKey: tbaAPIKey,
		certFile:  certFile,
		keyFile:   keyFile,
	}

	// setup report schema

	f, err := os.Open(schemaPath)
	if err != nil {
		return nil, err
	}

	s.schema = make(analysis.Schema)

	if err = json.NewDecoder(f).Decode(&s.schema); err != nil {
		return nil, err
	}

	// setup routes

	s.handler = corsMiddleware(s.newRouter())

	// setup jwt secret

	jwtSecret := make([]byte, 64)
	_, err = rand.Read(jwtSecret)
	if err != nil {
		return s, fmt.Errorf("generating jwt secret: %v", err)
	}

	return s, nil
}

// Run starts a running the server on the specified address
func (s *Server) Run(httpAddr, httpsAddr string) error {
	s.pollEvents()

	eventTicker := time.NewTicker(time.Hour * 24)
	defer eventTicker.Stop()
	go func() {
		for range eventTicker.C {
			s.pollEvents()
		}
	}()

	errChan := make(chan error)

	if s.certFile == "" || s.keyFile == "" {
		return http.ListenAndServe(httpAddr, s.handler)
	}

	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           s.handler,
		ReadTimeout:       time.Second * 15,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 15,
		IdleTimeout:       time.Second * 30,
		MaxHeaderBytes:    4096,
	}

	httpsServer := &http.Server{
		Addr:              httpsAddr,
		Handler:           s.handler,
		ReadTimeout:       time.Second * 15,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 15,
		IdleTimeout:       time.Second * 30,
		MaxHeaderBytes:    4096,
	}

	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	go func() {
		errChan <- httpsServer.ListenAndServeTLS(s.certFile, s.keyFile)
	}()

	err := <-errChan
	httpServer.Close()
	httpsServer.Close()

	return err
}

func (s *Server) newRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/authenticate", s.authenticateHandler)
	router.Handle("/users", s.authHandler(http.HandlerFunc(s.createUserHandler)))
	router.Handle("/users/{username}", s.authHandler(http.HandlerFunc(s.deleteUserHandler)))

	router.HandleFunc("/events", s.eventsHandler).Methods("GET")
	router.Handle("/events/{eventKey}", s.pollMatchMiddleware(http.HandlerFunc(s.eventHandler))).Methods("GET")
	router.Handle("/events/{eventKey}/{matchKey}", s.pollMatchMiddleware(http.HandlerFunc(s.matchHandler))).Methods("GET")

	router.Handle("/reports/{eventKey}/{matchKey}", s.authHandler(http.HandlerFunc(s.reportHandler))).Methods("PUT")

	router.HandleFunc("/schema", s.schemaHandler).Methods("GET")

	router.HandleFunc("/analysis/{eventKey}", s.eventAnalysisHandler).Methods("GET")
	router.HandleFunc("/analysis/{eventKey}/{team}", s.teamAnalysisHandler).Methods("GET")
	router.HandleFunc("/analysis/{eventKey}/{matchKey}/{color}", s.allianceAnalysisHandler).Methods("GET")

	return router
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

const tbaURL = "http://www.thebluealliance.com/api/v3"

func (s *Server) pollEvents() {
	bEvents, err := tba.GetEvents(tbaURL, s.tbaAPIKey, time.Now().Year())
	if err == tba.ErrNotModified {
		return
	} else if err != nil {
		s.logger.LogJSON(map[string]interface{}{"error": fmt.Errorf("server: polling events: %v", err).Error()})
		return
	}

	if err := s.store.Event.MassUpsert(bEvents); err != nil {
		s.logger.LogJSON(map[string]interface{}{"error": fmt.Errorf("server: updating events: %v", err).Error()})
	}
}

func (s *Server) pollMatches(eventKey string) {
	matches, err := tba.GetMatches(tbaURL, s.tbaAPIKey, eventKey)
	if err == tba.ErrNotModified {
		return
	} else if err != nil {
		s.logger.LogJSON(map[string]interface{}{"error": fmt.Errorf("server: polling matches for event '%s': %v", eventKey, err).Error()})
		return
	}

	if err := s.store.Match.MassUpsert(matches, s.store.Alliance); err != nil {
		s.logger.LogJSON(map[string]interface{}{"error": fmt.Errorf("server: updating matches for event '%s': %v", eventKey, err).Error()})
	}
}

func (s *Server) pollMatchMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.pollMatches(mux.Vars(r)["eventKey"])

		next.ServeHTTP(w, r)
	})
}
