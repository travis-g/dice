package server

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
	DebugMode             bool
	Port                  int
	PrettifyLogs          bool
	StaticContentDir      string

	// Regular Expressions
	DiceRegex      = regexp.MustCompile(`(?i)(?P<count>\d*)d(?P<size>\d+)`)
	ValidCalcRegex = regexp.MustCompile(`(?i)^((\d+)|(\d*d\d+))((\s*[\+-]\s*)((\d+)|(\d*d\d+)))*$`)
)

func Run() (int, error) {
	// Flag parsing
	flag.BoolVar(&DebugMode, "debug", false, "run the server in debug mode with higher verbosity")
	flag.BoolVar(&PrettifyLogs, "pretty", false, "prettify output logs. If false, outputs JSON logs")
	flag.IntVar(&Port, "port", 8000, "port to listen on")
	flag.Parse()

	// InfoLevel logging by default
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if PrettifyLogs {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Set up debug options if debugging
	if DebugMode {
		// Increase log verbosity
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("debug mode enabled")
	}

	// Seed random
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	log.Debug().Int64("seed", seed).Msg("seeded PRNG")

	// Configure routing
	r := ConfigureRouting()

	// Define the server
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + strconv.Itoa(Port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	// Run server as a goroutine so that we don't block
	// This lets us set up the interrupt signal handling
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("server fatal error")
		}
	}()

	log.Info().Str("address", srv.Addr).Msg("server started")

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
	return 0, nil
}

func main() {
	exit, err := Run()
	if err != nil {
		log.Error().Err(err).Msg("exited with error")
	}
	os.Exit(exit)
}
