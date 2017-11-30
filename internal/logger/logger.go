package logger

import (
	"io"
	"log"
	"net/http"
	"time"
)

const (
	infoPrefix  = "[INFO]"
	debugPrefix = "[DEBUG]"
	errorPrefix = "[ERROR]"
)

// Service is provided for writing log messages to an io.Writer interface
type Service struct {
	logger   *log.Logger
	settings Settings
}

// Settings holds the settings for what log messages to print
type Settings struct {
	Info  bool
	Debug bool
	Error bool
}

// New creates a new service for logging given an io.Writer to write log messages to
func New(out io.Writer, settings Settings) Service {
	return Service{logger: log.New(out, "", log.LstdFlags), settings: settings}
}

// Middleware wraps an HTTP handler to log information about
// the request such as the method, URI, name, and time to complete
func (s Service) Middleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		s.Debugf(
			"%-8s%s%d",
			r.Method,
			r.URL,
			time.Since(start),
		)
	})
}

// Infof will print to the io.Writer if Info is enabled with the
// prefix '[INFO]'
func (s Service) Infof(format string, a ...interface{}) {
	if s.settings.Info {
		s.logger.Printf(infoPrefix+" "+format, a...)
	}
}

// Debugf will print to the io.Writer if Debug is enabled with the
// prefix '[DEBUG]'
func (s Service) Debugf(format string, a ...interface{}) {
	if s.settings.Debug {
		s.logger.Printf(debugPrefix+" "+format, a...)
	}
}

// Errorf will print to the io.Writer if Error is enabled with the
// prefix '[ERROR]'
func (s Service) Errorf(format string, a ...interface{}) {
	if s.settings.Error {
		s.logger.Printf(errorPrefix+" "+format, a...)
	}
}