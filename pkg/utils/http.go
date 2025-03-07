package utils

import (
	"encoding/json"
	"net/http"
)

// RespondJSON sends a JSON response with the given status code
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		return json.NewEncoder(w).Encode(data)
	}

	return nil
}

// RespondError sends an error response with the given status code
func RespondError(w http.ResponseWriter, statusCode int, message string) error {
	return RespondJSON(w, statusCode, map[string]string{"error": message})
}

// ParseJSON parses a JSON request body into the given struct
func ParseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
