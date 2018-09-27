package server

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	response, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, err string) {
	respondWithJSON(w, code, map[string]string{
		"error": err,
	})
}

// isValidDiceNotationString returns if a given string is a valid dice notation
// expression. If debug mode is enabled all strings are matched.
func isValidDiceNotationString(s string) bool {
	if DebugMode {
		return true
	}
	return ValidCalcRegex.Match([]byte(s))
}

func HandleResponse(w http.ResponseWriter, jsonString string) {
	var f map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &f)
	if err != nil {
		log.Error().Err(err).Msg("json error")
	}
	respondWithJSON(w, http.StatusOK, jsonString)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotFound, "not found")
}

func RollRequestHandler() {
	// TODO this function should handle all Die rolling requests
}

func RollHandler(w http.ResponseWriter, r *http.Request) {
	// Grab the dice notation string from the request URI
	roll := mux.Vars(r)["roll"]

	if !isValidDiceNotationString(roll) {
		respondWithError(w, http.StatusBadRequest, "invalid dice string")
		return
	}

	result, err := EvalDiceNotationString(roll)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "bad dice notation string")
		return
	}
	respondWithJSON(w, http.StatusOK, result)
}

// RootHandler handles requests to the base server. This should be replaced with
// an API description or static HTML page.
func RootHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"prompt": "You approach the server.",
	})
}

// SysToolsRandomHandler returns a random integer result of rand.Int().
func ToolsRandomHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"result": rand.Int(),
	})
}
