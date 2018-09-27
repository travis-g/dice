package server

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, _ := url.PathUnescape(r.RequestURI)
		log.Info().
			Str("method", r.Method).
			Str("path", path)
		// Call the actual request handler
		next.ServeHTTP(w, r)
	})
}

func ConfigureRouting() *mux.Router {
	// Configure routing
	r := mux.NewRouter()

	// Send every request through logging middleware for statistics
	r.Use(loggingMiddleware)

	r.HandleFunc("/", RootHandler)

	// Subrouting for actual API stuff, which has a /v1/ prefix
	s := r.PathPrefix("/v1").Subrouter()
	s.HandleFunc("/roll/{roll}", RollHandler)

	// Default roll
	r.HandleFunc("/{roll}", RollHandler)

	return r
}
