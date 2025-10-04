package middleware

import (
	"encoding/json"
	"log"
	"net/http"
)

const validAPIKey = "secret123"

type ErrorResponse struct {
	Error string `json:"error"`
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != validAPIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
			return
		}

		next(w, r)
	}
}