package server

import (
	"net/http"

	"github.com/Pigmice2733/scouting-backend/internal/mroute"
	"github.com/gorilla/mux"
)

func initRoutes(r *mux.Router, s *Server) {
	mroute.HandleRoutes(r, map[string]mroute.Route{
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
}
