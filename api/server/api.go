package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/fgrimme/patrongg/store"
	"github.com/gorilla/mux"
)

// HTTP errors
var (
	errInternal   = errors.New("internal_error")
	errNotFound   = errors.New("not_found")
	errBadRequest = errors.New("bad_request")
)

const (
	Active  = "active"
	Benched = "benched"
)

// rosterStore handles operations on rosters.
type rosterStore interface {
	Get(ctx context.Context, rosterID uint64) (*store.Roster, error)
}

// rosterService provides API methods to operate on rosters.
type rosterService struct {
	rosterStore
	timeout time.Duration
}

// ServeHTTP serves requests to the roster enpoint.
func (rs *rosterService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), rs.timeout)
	defer cancel()
	// we attach the logger from the request to the context so we do need
	// to pass it as an parameter
	ctx = loggerFromRequest(r).WithContext(ctx)

	// query param validation is currently performed by mux only
	vars := mux.Vars(r)
	rosterID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		// note, this is non-reachable code whith the current mux routing setup
		writeError(w, r, errBadRequest, http.StatusBadRequest)
		return
	}

	if status, ok := vars["status"]; !ok {
		rs.getRoster(ctx, w, r, rosterID)
		return
	} else if status == Active {
		rs.getPlayers(ctx, w, r, rosterID, Active)
		return
	} else if status == "benched" {
		rs.getPlayers(ctx, w, r, rosterID, Benched)
		return
	}

	// note, this is non-reachable code whith the current mux routing setup
	writeError(w, r, errNotFound, http.StatusNotFound)
	return
}

// getRoster responds with a representation of the entire roster for the given
// id or an error.
func (rs *rosterService) getRoster(ctx context.Context, w http.ResponseWriter, r *http.Request, rosterID uint64) {
	players, err := rs.Get(ctx, rosterID)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}
	encodeJSON(w, r, players, http.StatusOK)
	return
}

// getPlayers responds with a representation of the players with the given status
// of the roster with the given id or an error.
func (rs *rosterService) getPlayers(ctx context.Context, w http.ResponseWriter, r *http.Request, rosterID uint64, status string) {
	roster, err := rs.Get(ctx, rosterID)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}
	if status == Active {
		encodeJSON(w, r, roster.Players.Active, http.StatusOK)
		return
	}
	if status == Benched {
		encodeJSON(w, r, roster.Players.Benched, http.StatusOK)
		return
	}
	writeError(w, r, errNotFound, http.StatusNotFound)
}

// playerStore provides methods to operate on the players store.
type playerStore interface {
	Insert(ctx context.Context, player store.Player) (*store.Player, error)
	Update(ctx context.Context, player store.Player) (*store.Player, error)
	ChangePlayers(ctx context.Context, players store.PlayerChange) (*store.PlayerChange, error)
}

type playerService struct {
	playerStore
	timeout time.Duration
}

// ServeHTTP serves requests to the players enpoint.
func (ps *playerService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// we attach the logger from the request to the context so we do need
	// to pass it as an parameter
	ctx, cancel := context.WithTimeout(r.Context(), ps.timeout)
	defer cancel()
	ctx = loggerFromRequest(r).WithContext(ctx)

	// add a player
	// new players are benched by default
	if r.Method == http.MethodPost {
		// we expect a request body that represents a player or we consider
		// the request as invalid
		var player store.Player
		if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
			writeError(w, r, errBadRequest, http.StatusBadRequest)
			return
		}
		player.Status = Benched // benched by default
		ps.insert(ctx, w, r, player)
		return
	}

	// modify a player
	if r.Method == http.MethodPatch {
		_, route := path.Split(r.URL.Path)
		switch route {
		case "update":
			// we expect a request body that represents a player with the new
			// rosterID or we consider the request as invalid
			var player store.Player
			if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
				writeError(w, r, err, http.StatusInternalServerError)
				return
			}
			// here we chose which fields we want to update
			// players always get benched when they are added to a roster
			rosterUpdate := store.Player{
				PlayerID: player.PlayerID,
				RosterID: player.RosterID,
				Status:   Benched,
			}
			ps.update(ctx, w, r, rosterUpdate)
			return

		case "change":
			// we expect a request body that contains two players or we
			// consider the request as invalid
			var players store.PlayerChange
			if err := json.NewDecoder(r.Body).Decode(&players); err != nil {
				writeError(w, r, err, http.StatusInternalServerError)
				return
			}
			ps.change(ctx, w, r, players)
			return
		}
	}

	// note, this is non-reachable code whith the current mux routing setup
	writeError(w, r, errNotFound, http.StatusNotFound)
}

// insert inserts a new player to the datastore. Responds with the newly created
// player with a generated player id or an error (and thus is POST compliant).
func (ps *playerService) insert(ctx context.Context, w http.ResponseWriter, r *http.Request, player store.Player) {
	p, err := ps.Insert(ctx, player)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}
	encodeJSON(w, r, p, http.StatusOK)
	return
}

// update updates the given player. Responds with the updated/patched player or
// an error (and thus is HTTP/PATCH compliant).
func (ps *playerService) update(ctx context.Context, w http.ResponseWriter, r *http.Request, player store.Player) {
	p, err := ps.Update(ctx, player)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}
	encodeJSON(w, r, p, http.StatusOK)
	return
}

// change swaps two players statuses. Responds the updated/patched
// players or an error (and thus is HTTP/PATCH compliant).
func (ps *playerService) change(ctx context.Context, w http.ResponseWriter, r *http.Request, players store.PlayerChange) {
	p, err := ps.ChangePlayers(ctx, players)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}
	encodeJSON(w, r, p, http.StatusOK)
	return
}
