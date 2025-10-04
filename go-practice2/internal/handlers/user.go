package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type UserGetResponse struct {
	UserID int `json:"user_id"`
}

type UserPostRequest struct {
	Name string `json:"name"`
}

type UserPostResponse struct {
	Created string `json:"created"`
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetUser(w, r)
	case http.MethodPost:
		handlePostUser(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	respondWithJSON(w, http.StatusOK, UserGetResponse{UserID: id})
}

func handlePostUser(w http.ResponseWriter, r *http.Request) {
	var req UserPostRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid name")
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(req.Name) == "" {
		respondWithError(w, http.StatusBadRequest, "invalid name")
		return
	}

	respondWithJSON(w, http.StatusCreated, UserPostResponse{Created: req.Name})
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, ErrorResponse{Error: message})
}