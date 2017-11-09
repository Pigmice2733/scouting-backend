package logger

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Service is provided for writing log messages to an io.Writer interface
type Service struct {
	logger *log.Logger
}

// New creates a new service for logging given an io.Writer to write log messages to
func New(out io.Writer) Service {
	return Service{logger: log.New(out, "", log.LstdFlags)}
}

// Middleware wraps an HTTP handler to log information about
// the request such as the method, URI, name, and time to complete
func (s Service) Middleware(inner http.HandlerFunc, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		if s.logger == nil {
			s.logger = log.New(os.Stdout, "", log.LstdFlags)
		}

		s.logger.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
