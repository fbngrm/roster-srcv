package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fgrimme/patrongg/api/server"
	"github.com/fgrimme/patrongg/database"
	"github.com/fgrimme/patrongg/store/player"
	"github.com/fgrimme/patrongg/store/roster"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unkown" // version is build into the binary, see Makefile

	// provide the configuration via env parameters or arguments
	httpAddr      = kingpin.Flag("http-addr", "address of HTTP server").Envar("HTTP_ADDR").Required().String()
	serviceName   = kingpin.Flag("service", "service name").Envar("SERVICE").Default("roster-service").String()
	playerDBDSN   = kingpin.Flag("player-db-dsn", "player db DSN").Envar("PLAYER_DB_DSN").Required().String()
	timeout       = kingpin.Flag("timeout", "timeout to handle incoming requests").Envar("REQ_TIMEOUT").Default("900ms").Duration()
	shutdownDelay = kingpin.Flag("shutdown-delay", "shutdown delay in ms").Envar("SHUTDOWN_DELAY").Default("5000ms").Duration()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	// we use the default log level debug and write to stderr.
	// note, we log in (inefficient) human friendly format to console here since it
	// is a coding challenge. in a production environment we would prefer structured,
	// machine parsable format so we could make use of automated log analysis.
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// replace standard log
	log.SetFlags(0)
	log.SetOutput(logger)
	logger = logger.With().
		Interface("service", serviceName).
		Interface("version", version).
		Logger()

	// connect to databases
	ds, err := database.Connect("postgres", *playerDBDSN, "player_db", *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *serviceName, err)
		os.Exit(1)
	}
	defer func() {
		if err := ds.Close(); err != nil {
			logger.Warn().Msgf("cleaning up DB resources: %v", err)
		}
	}()

	// we use dependency injection throughout the whole application to either create
	// working instances or fail early on instantiation
	httpSrv, err := server.New(*httpAddr, *timeout, roster.New(ds), player.New(ds), logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", *serviceName, err)
		os.Exit(1)
	}

	// run and handle shutdown gracefully
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		<-quit
		cancel()
	}()

	go httpSrv.Run()

	<-ctx.Done()

	server.HealthCheckShutDown()

	// set a shutdown deadline
	ctx, cancel = context.WithTimeout(context.Background(), *shutdownDelay)
	defer cancel()

	// when shutting down, we first gracefully shutting down the main http
	// server, waiting for it to finish processing all the running requests.
	httpSrv.Shutdown(ctx)
	// here we would also shut down the metrics server so we would set the
	// shutdown timeout to something higher than the prometheus scrape interval.
	// we would use a counter to check if a scrape happened to shutdown
	// asap though
}
