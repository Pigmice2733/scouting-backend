package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Pigmice2733/scouting-backend/internal/mroute"
	"github.com/Pigmice2733/scouting-backend/internal/server/logic"
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
	consumer  tba.Consumer
	logger    logger.Service
	schema    analysis.Schema
	jwtSecret []byte
	certFile  string
	keyFile   string
}

// New creates a new server given a db file and a io.Writer for logging
func New(store *store.Service, consumer tba.Consumer, logWriter io.Writer, schemaPath string, certFile, keyFile string) (*Server, error) {
	s := &Server{
		logger:   logger.New(logWriter),
		store:    store,
		consumer: consumer,
		certFile: certFile,
		keyFile:  keyFile,
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

	s.handler = cors(limitBody(s.newRouter()))

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

	mroute.HandleRoutes(router, map[string]mroute.Route{
		"/authenticate": mroute.Simple(http.HandlerFunc(s.authenticateHandler), "POST"),

		"/users": {
			Handler: mroute.Multi(map[string]http.Handler{
				"GET":  http.HandlerFunc(s.getUsersHandler),
				"POST": http.HandlerFunc(s.createUserHandler),
			}),
			Methods:     []string{"GET", "POST"},
			Middlewares: []mroute.Middleware{s.authHandler, adminHandler},
		},

		"/users/{username}": {
			Handler: mroute.Multi(map[string]http.Handler{
				"POST":   http.HandlerFunc(s.updateUserHandler),
				"DELETE": http.HandlerFunc(s.deleteUserHandler),
			}),
			Methods:     []string{"POST", "DELETE"},
			Middlewares: []mroute.Middleware{s.authHandler},
		},

		"/events":                       mroute.Simple(http.HandlerFunc(s.eventsHandler), "GET", cache),
		"/events/{eventKey}":            mroute.Simple(http.HandlerFunc(s.eventHandler), "GET", cache, s.pollMatchMiddleware),
		"/events/{eventKey}/{matchKey}": mroute.Simple(http.HandlerFunc(s.matchHandler), "GET", cache, s.pollMatchMiddleware),

		"/reports/{eventKey}/{matchKey}": mroute.Simple(http.HandlerFunc(s.reportHandler), "PUT", cache, s.authHandler),

		"/schema": mroute.Simple(http.HandlerFunc(s.schemaHandler), "GET", cache),

		"/photo/{team}": mroute.Simple(http.HandlerFunc(s.photoHandler), "GET", cache),

		"/analysis/{eventKey}":                    mroute.Simple(http.HandlerFunc(s.eventAnalysisHandler), "GET"),
		"/analysis/{eventKey}/{team}":             mroute.Simple(http.HandlerFunc(s.teamAnalysisHandler), "GET"),
		"/analysis/{eventKey}/{matchKey}/{color}": mroute.Simple(http.HandlerFunc(s.allianceAnalysisHandler), "GET"),
	})

	return router
}

func (s *Server) photoHandler(w http.ResponseWriter, r *http.Request) {
	team := mux.Vars(r)["team"]

	url, err := logic.GetPhoto(team, time.Now().Year(), s.store.Photo, s.consumer)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		s.logger.LogRequestError(r, fmt.Errorf("getting team photo: %v", err))
		return
	}

	if url == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	resp, err := http.Get(url)
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		s.logger.LogRequestError(r, fmt.Errorf("getting team media: %v", err))
		return
	}
	defer resp.Body.Close()

	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		s.logger.LogRequestError(r, fmt.Errorf("copying team media to client: %v", err))
		return
	}
}

func (s *Server) pollEvents() {
	bEvents, err := s.consumer.GetEvents(time.Now().Year())
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
	matches, err := s.consumer.GetMatches(eventKey)
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
