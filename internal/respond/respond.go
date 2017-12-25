package respond

import (
	"encoding/json"
	"net/http"
)

// JSON responds with the first non-nil payload, formats error messages
func JSON(w http.ResponseWriter, responses ...interface{}) {
	respond := func(payload interface{}) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
			respond(map[string]string{"error": err.Error()})
		case error:
			respond(map[string]string{"error": value.Error()})
		default:
			respond(response)
		}
		// exit on the first output
		break
	}
}
