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
	"github.com/Pigmice2733/scouting-backend/internal/respond"
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
	year      int
}

// New creates a new server given a db file and a io.Writer for logging
func New(store *store.Service, consumer tba.Consumer, logWriter io.Writer, year int, origin, schemaPath string, certFile, keyFile string) (*Server, error) {
	s := &Server{
		logger:   logger.New(logWriter),
		store:    store,
		consumer: consumer,
		certFile: certFile,
		keyFile:  keyFile,
		year:     year,
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

	s.handler = s.newHandler(origin)

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

func (s *Server) newHandler(origin string) http.Handler {
	router := mux.NewRouter()

	mroute.HandleRoutes(router, map[string]mroute.Route{
		"/authenticate": mroute.Simple(http.HandlerFunc(s.authenticateHandler), "POST"),

		"/users": {
			Handler: mroute.Multi(map[string]http.Handler{
				"GET":  s.authHandler(adminHandler(http.HandlerFunc(s.getUsersHandler))),
				"POST": http.HandlerFunc(s.createUserHandler),
			}),
			Methods:     []string{"GET", "POST"},
			Middlewares: []mroute.Middleware{},
		},

		"/users/{username}": {
			Handler: mroute.Multi(map[string]http.Handler{
				"PUT":    http.HandlerFunc(s.updateUserHandler),
				"DELETE": http.HandlerFunc(s.deleteUserHandler),
			}),
			Methods:     []string{"PUT", "DELETE"},
			Middlewares: []mroute.Middleware{s.authHandler},
		},

		"/events":                               mroute.Simple(http.HandlerFunc(s.eventsHandler), "GET", cache),
		"/events/{eventKey}":                    mroute.Simple(http.HandlerFunc(s.eventHandler), "GET", cache, s.pollMatchMiddleware),
		"/events/{eventKey}/teams":              mroute.Simple(http.HandlerFunc(s.teamsAtEventHandler), "GET", cache),
		"/events/{eventKey}/matches/{matchKey}": mroute.Simple(http.HandlerFunc(s.matchHandler), "GET", cache, s.pollMatchMiddleware),

		"/events/{eventKey}/matches/{matchKey}/reports": mroute.Simple(http.HandlerFunc(s.reportHandler), "PUT", s.authHandler),
		"/events/{eventKey}/teams/{team}/reports":       mroute.Simple(http.HandlerFunc(s.getTeamEventReportsHandler), "GET", cache, s.pollMatchMiddleware),
		"/teams/{team}/reports":                         mroute.Simple(http.HandlerFunc(s.getTeamReportsHandler), "GET", cache, s.pollMatchMiddleware),

		"/schema": mroute.Simple(http.HandlerFunc(s.schemaHandler), "GET", cache),

		"/photo/{team}": mroute.Simple(http.HandlerFunc(s.photoHandler), "GET", cache),

		"/events/{eventKey}/analysis":                                     mroute.Simple(http.HandlerFunc(s.eventAnalysisHandler), "GET", s.pollMatchMiddleware),
		"/events/{eventKey}/teams/{team}/analysis":                        mroute.Simple(http.HandlerFunc(s.teamAnalysisHandler), "GET", s.pollMatchMiddleware),
		"/events/{eventKey}/matches/{matchKey}/alliance/{color}/analysis": mroute.Simple(http.HandlerFunc(s.allianceAnalysisHandler), "GET", s.pollMatchMiddleware),

		"/picklists": {
			Handler: mroute.Multi(map[string]http.Handler{
				"GET":  http.HandlerFunc(s.picklistsHandler),
				"POST": http.HandlerFunc(s.newPicklistHandler),
			}),
			Methods:     []string{"GET", "POST"},
			Middlewares: []mroute.Middleware{s.authHandler},
		},
		"/picklists/{id}": {
			Handler: mroute.Multi(map[string]http.Handler{
				"GET":    http.HandlerFunc(s.picklistHandler),
				"PUT":    s.authHandler(http.HandlerFunc(s.updatePicklistHandler)),
				"DELETE": s.authHandler(http.HandlerFunc(s.deletePicklistHandler)),
			}),
			Methods: []string{"GET", "PUT", "DELETE"},
		},
		"/picklists/event/{eventKey}": mroute.Simple(http.HandlerFunc(s.picklistEventHandler), "GET", s.authHandler),

		"/leaderboard": mroute.Simple(http.HandlerFunc(s.leaderboardHandler), "GET"),
	})

	return cors(limitBody(router), origin)
}

func (s *Server) teamsAtEventHandler(w http.ResponseWriter, r *http.Request) {
	eventKey := mux.Vars(r)["eventKey"]

	teams, err := s.store.Report.GetReportedOn(eventKey)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		s.logger.LogRequestError(r, fmt.Errorf("getting reported on: %v", err))
		return
	}

	respond.JSON(w, teams)
}

func (s *Server) photoHandler(w http.ResponseWriter, r *http.Request) {
	team := mux.Vars(r)["team"]

	url, err := logic.GetPhoto(team, s.year, s.store.Photo, s.consumer)
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

func (s *Server) leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	type stat struct {
		Reporter string `json:"reporter"`
		Reports  int    `json:"reports"`
	}

	var resp []stat

	stats, err := s.store.Report.GetReporterStats()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		s.logger.LogRequestError(r, fmt.Errorf("getting reporter stats: %v", err))
		return
	}

	for reporter, reports := range stats {
		resp = append(resp, stat{reporter, reports})
	}

	respond.JSON(w, resp)
}

func (s *Server) pollEvents() {
	bEvents, err := s.consumer.GetEvents(s.year)
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
		eventKey := mux.Vars(r)["eventKey"]

		if eventKey != "" {
			s.pollMatches(eventKey)
		}

		next.ServeHTTP(w, r)
	})
}
