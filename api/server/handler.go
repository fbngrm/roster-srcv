package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fgrimme/patrongg/api"
	"github.com/fgrimme/patrongg/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// newHandler creates an http handler that operates on rosters and players.
func newHandler(rs rosterStore, ps playerStore, timeout time.Duration, logger zerolog.Logger) (http.Handler, error) {
	var mw []middleware.Middleware
	mw = append(mw, middleware.NewRecoverHandler())
	mw = append(mw, middleware.NewContextLog(logger)...)

	// services handle http requests and hold a store to operate on a database
	rosterSrvc := middleware.Use(&rosterService{rs, timeout}, mw...)
	playerSrvc := middleware.Use(&playerService{ps, timeout}, mw...)

	router := mux.NewRouter()
	router.Handle("/ready", &readinessHandler{}).Methods("GET")

	// roster store
	router.Handle("/roster/{id:[0-9]+}", rosterSrvc).Methods("GET")
	router.Handle(fmt.Sprintf("/roster/{id:[0-9]+}/{status:(?:%s|%s)}", Active, Benched), rosterSrvc).Methods("GET")

	// player store
	router.Handle("/players/add", playerSrvc).Methods("POST")
	router.Handle("/players/update", playerSrvc).Methods("PATCH")
	router.Handle("/players/change", playerSrvc).Methods("PATCH")

	return router, nil
}

// encodeJSON encodes v to w in JSON format.
func encodeJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		loggerFromRequest(r).Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
	}
}

func loggerFromRequest(r *http.Request) *zerolog.Logger {
	logger := hlog.FromRequest(r).With().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Logger()
	return &logger
}

// writeError writes an error to the http response in JSON format.
func writeError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// prepare log
	logger := loggerFromRequest(r).With().
		Err(err).
		Int("status", code).
		Logger()
	// hide error from client if it's internal
	if code == http.StatusInternalServerError {
		logger.Error().Msg("unexpected http error")
		err = errInternal
	} else {
		logger.Debug().Msg("http error")
	}
	encodeJSON(w, r, &api.Error{Err: err.Error()}, code)
}
