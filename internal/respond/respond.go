package respond

import (
	"encoding/json"
	"net/http"
)

// Error returns a structured error for API responses
func Error(err error) interface{} {
	response := struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}{}

	response.Error.Message = err.Error()

	return response
}

// JSON responds with the first non-nil payload, formats error messages
func JSON(w http.ResponseWriter, responses ...interface{}) {
	respond := func(payload interface{}) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	for _, response := range responses {
		switch value := response.(type) {
		case nil:
			continue
		case func() error:
			err := value()
			if err == nil {
				continue
			}
			respond(Error(err))
		case error:
			respond(Error(value))
		default:
			respond(struct {
				Response interface{} `json:"response"`
			}{response})
		}
		// exit on the first output
		break
	}
}
