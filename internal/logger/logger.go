package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Service is a struct for logging to a specified writer.
type Service struct {
	out         io.Writer
	jsonEncoder *json.Encoder
}

// New creates a new logging service for logging to the specified output.
func New(out io.Writer) Service {
	return Service{out: out, jsonEncoder: json.NewEncoder(out)}
}

// Logf allows you to log to output with printf-style logging. Prefer
// structured logging (LogJSON) to this.
func (l *Service) Logf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(l.out, format, a...)
}

// LogJSON allows you to make strucutured logs written to the output after
// being json encoded. Prefer this to non structured logging (Logf).
func (l *Service) LogJSON(info map[string]interface{}) error {
	return l.jsonEncoder.Encode(info)
}

// LogRequestJSON makes json encoded structured logs by using LogJSON, but
// adds information about a request (path, method) to the log.
func (l *Service) LogRequestJSON(r *http.Request, inf map[string]interface{}) error {
	reqLog := map[string]interface{}{"path": r.URL.Path, "method": r.Method}
	for k, v := range inf {
		reqLog[k] = v
	}
	return l.LogJSON(reqLog)
}
