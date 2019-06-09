package command

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/travis-g/dice"
	"github.com/urfave/cli"
)

// handleRequest handles a POST request that supplies a properties list by
// creating a new die, rolling it, and returning the rolled die and any errors.
func handleRequest(ctx context.Context, props dice.RollerProperties) (roll dice.Roller, err error) {
	roll, err = dice.NewRoller(&props)
	if err != nil {
		return
	}
	err = roll.Roll(ctx)
	return
}

func rollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	props, err := dice.ParseNotation(vars["roll"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	group, err := dice.NewGroup(props)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = group.Roll(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"dice": group,
	}

	x, err := toJSON(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(x))
}

func rollPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	decoder := json.NewDecoder(r.Body)
	var vars map[string]interface{}
	err := decoder.Decode(&vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	props, err := dice.ParseNotation(vars["roll"].(string))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	group, err := dice.NewGroup(props)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = group.Roll(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"dice": group,
	}

	x, err := toJSON(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(x))
}

// ServerCommand is a command that will initialize a DRAAS HTTP server.
//
// The server routines themselves should be split into a separate dice/server
// package.
func ServerCommand(c *cli.Context) error {
	r := mux.NewRouter()

	r.HandleFunc("/roll/{roll}", rollHandler).Methods("GET")
	r.HandleFunc("/roll", rollPostHandler).Methods("POST")

	srv := &http.Server{
		Addr:         c.String("http"),
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 10,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(sig, os.Interrupt)

	<-sig

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("shutting down")
	return nil
}
