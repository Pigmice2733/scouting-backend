package mroute

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Middleware functions wrap an http handler.
type Middleware func(http.Handler) http.Handler

// Route holds information for an HTTP route.
type Route struct {
	Handler     http.Handler
	Methods     []string
	Middlewares []Middleware
}

// Multi allows you to specify what handler should be used depending on
// the method. Note that a panic will occur if a request is given for a method
// that does not have a route.
func Multi(routes map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routes[r.Method].ServeHTTP(w, r)
	})
}

// Simple is a helper function that creates a route for just a single method.
func Simple(h http.Handler, method string, Middlewares ...Middleware) Route {
	return Route{Handler: h, Methods: []string{method}, Middlewares: Middlewares}
}

// HandleRoutes applies routes to a mux router. OPTIONS requests are handled,
// and the Access-Control-Allowed-Methods header is set. Middlewares are applied
// from the inside out (ex: cors, cache, auth would apply auth, then cors, then
// cache, then the handler).
func HandleRoutes(r *mux.Router, routes map[string]Route) {
	for path, route := range routes {
		r.Handle(path, routeHandler(route))
	}
}

func routeHandler(route Route) http.Handler {
	for i := len(route.Middlewares) - 1; i >= 0; i-- {
		route.Handler = route.Middlewares[i](route.Handler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods := append(route.Methods, "OPTIONS")

		if !existsIn(r.Method, methods) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))

		if r.Method == "OPTIONS" {
			return
		}

		route.Handler.ServeHTTP(w, r)
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
