package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/Pigmice2733/scouting-backend/internal/tba"
	"github.com/fharding1/ezetag"

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

	s.handler = s.newRouter()

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

	router.Handle("/authenticate", cors(http.HandlerFunc(s.authenticateHandler), []string{"POST"}))
	router.Handle("/users", cors(s.authHandler(http.HandlerFunc(s.createUserHandler)), []string{"POST"}))
	router.Handle("/users/{username}", cors(s.authHandler(http.HandlerFunc(s.deleteUserHandler)), []string{"DELETE"}))

	router.Handle("/events", stdMiddleware(http.HandlerFunc(s.eventsHandler)))
	router.Handle("/events/{eventKey}", stdMiddleware(s.pollMatchMiddleware(http.HandlerFunc(s.eventHandler))))
	router.Handle("/events/{eventKey}/{matchKey}", stdMiddleware(s.pollMatchMiddleware(http.HandlerFunc(s.matchHandler))))

	router.Handle("/reports/{eventKey}/{matchKey}", cors(s.authHandler(http.HandlerFunc(s.reportHandler)), []string{"PUT"}))

	router.Handle("/schema", stdMiddleware(http.HandlerFunc(s.schemaHandler)))

	router.Handle("/analysis/{eventKey}", stdMiddleware(http.HandlerFunc(s.eventAnalysisHandler)))
	router.Handle("/analysis/{eventKey}/{team}", stdMiddleware(http.HandlerFunc(s.teamAnalysisHandler)))
	router.Handle("/analysis/{eventKey}/{matchKey}/{color}", stdMiddleware(http.HandlerFunc(s.allianceAnalysisHandler)))

	return router
}

func cors(next http.Handler, allowedMethods []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(append(allowedMethods, "OPTIONS"), ","))

		if r.Method == "OPTIONS" {
			return
		} else if !existsIn(r.Method, allowedMethods) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func existsIn(str string, strs []string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

func cache(next http.Handler) http.Handler {
	return gziphandler.GzipHandler(ezetag.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=180") // 3 minute max age, overriden by /events

		next.ServeHTTP(w, r)
	}), sha256.New))
}

func stdMiddleware(next http.Handler) http.Handler {
	return cors(cache(next), []string{"GET"})
}

func (s *Server) pollEvents() {
	bEvents, err := tba.GetEvents(s.tbaAPIKey, time.Now().Year())
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
	matches, err := tba.GetMatches(s.tbaAPIKey, eventKey)
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
