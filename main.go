package main

import (
	"context"
	"flag"
	rand "math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	MaxDice               = 1000
	ShutdownGraceDuration = time.Second * 5
	Debug                 bool
	Port                  int
	StaticContentDir      string

	// Regular Expressions
	DiceRegex      = regexp.MustCompile(`(?i)(?P<count>\d*)d(?P<size>\d+)`)
	ValidCalcRegex = regexp.MustCompile(`(?i)^((\d+)|(\d*d\d+))((\s*[\+-]\s*)((\d+)|(\d*d\d+)))*$`)
)

func main() {
	// Flag parsing
	flag.BoolVar(&Debug, "debug", false, "whether to run the server in debug mode")
	flag.IntVar(&Port, "port", 8000, "port to host the server")
	flag.Parse()

	// InfoLevel logging by default
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Set up debug options if debugging
	if Debug {
		// Dev log
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("debug mode enabled")
	}

	// Seed random
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	log.Debug().Int64("seed", seed).Msg("seeded PRNG")

	// Configure routing
	r := ConfigureRouting()

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + strconv.Itoa(Port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	// Run server as a goroutine so we don't block
	// This lets us set up the interrupt signal handling
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("server fatal error")
		}
	}()

	log.Info().Msg("server started")

	// Graceful shutdowns when quit via SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Waits until SIGINT sent
	<-c
	log.Info().Msg("SIGINT received")

	ctx, cancel := context.WithTimeout(context.Background(), ShutdownGraceDuration)
	defer cancel()
	srv.Shutdown(ctx)
	log.Info().Msg("shutting down")
	os.Exit(0)
}
