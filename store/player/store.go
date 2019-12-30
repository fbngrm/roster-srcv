package player

import (
	"context"

	"github.com/fgrimme/patrongg/database"
	"github.com/fgrimme/patrongg/store"
)

// PlayerStore handles operations on the the
// players table of the encapsulated datastore.
type PlayerStore struct {
	db *database.DB
}

func New(db *database.DB) *PlayerStore {
	return &PlayerStore{
		db: db,
	}
}

// Insert inserts a new player to the datastore. The player id is not inserted
// and must be created by the datastore. Returns the newly created player with
// the generated id.
func (ps *PlayerStore) Insert(ctx context.Context, player store.Player) (*store.Player, error) {
	query := `
  INSERT INTO players(roster_id,first_name,last_name,alias,status)
  VALUES($1,$2,$3,$4,$5)
  RETURNING *
  `
	db := ps.db.GetDB()
	ctx, cancel := ps.db.RequestContext(ctx)
	defer cancel()

	var p store.Player
	err := db.QueryRowContext(ctx, query,
		player.RosterID,
		player.FirstName,
		player.LastName,
		player.Alias,
		player.Status).
		Scan(
			&p.PlayerID,
			&p.RosterID,
			&p.FirstName,
			&p.LastName,
			&p.Alias,
			&p.Status)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Update updates all non-empty/not-null fields of the given player except for
// the player_id. Fails if foreign-key constraint roster_id is violated e.g. a
// roster with the given id does not exists.
// Returns the updated/patched player.
func (ps *PlayerStore) Update(ctx context.Context, player store.Player) (*store.Player, error) {
	query := `
  UPDATE players
  SET
    roster_id  = COALESCE(NULLIF($2, CAST(0 AS BIGINT)), players.roster_id),
    first_name = COALESCE(NULLIF($3, ''), players.first_name),
    last_name  = COALESCE(NULLIF($4, ''), players.last_name),
    alias      = COALESCE(NULLIF($5, ''), players.alias),
	status     = COALESCE(NULLIF($6, ''), players.status)
  WHERE id = $1
  RETURNING *`

	db := ps.db.GetDB()
	ctx, cancel := ps.db.RequestContext(ctx)
	defer cancel()

	var p store.Player
	err := db.QueryRowContext(ctx, query,
		player.PlayerID,
		player.RosterID,
		player.FirstName,
		player.LastName,
		player.Alias,
		player.Status).
		Scan(
			&p.PlayerID,
			&p.RosterID,
			&p.FirstName,
			&p.LastName,
			&p.Alias,
			&p.Status)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ChangePlayers uses a transaction to swap two players statuses. This ensures,
// that the roster stays in a consistent state of no more and no less than the
// currently active players.
// In other words, it ensures that when a player gets moved from the bench to
// the active roster, an active player gets moved to the bench. The function
// succeeds only if the given player to activate is currently benched, the given
// player to be benched is currently active and both players are members of the
// same roster.
// In case of failure, the transaction is rolled back.
// Returns the updated/patched players.
func (ps *PlayerStore) ChangePlayers(ctx context.Context, players store.PlayerChange) (*store.PlayerChange, error) {
	query := `
  UPDATE players
  SET status = $1
  WHERE id = $2
  AND status = $3
  AND roster_id = ( -- ensures that both players are in the same roster
    SELECT roster_id
    FROM players
    WHERE id = $4
  )
  RETURNING *`

	db := ps.db.GetDB()
	ctx, cancel := ps.db.RequestContext(ctx)
	defer cancel()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer stmt.Close()

	var playerID uint64
	var rosterID uint64
	var firstName string
	var lastName string
	var alias string
	var status string

	// activate benched player
	err = stmt.QueryRowContext(
		ctx,
		"active", // new status
		players.Benched.PlayerID,
		"benched", // needs to be benched currently
		players.Active.PlayerID).
		Scan(
			&playerID,
			&rosterID,
			&firstName,
			&lastName,
			&alias,
			&status)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// newly actived player
	active := store.Player{
		PlayerID:  playerID,
		RosterID:  rosterID,
		FirstName: firstName,
		LastName:  lastName,
		Alias:     alias,
		Status:    status,
	}

	// bench active player
	err = stmt.QueryRowContext(
		ctx,
		"benched", // new status
		players.Active.PlayerID,
		"active", // needs to be active currently
		players.Benched.PlayerID).
		Scan(
			&playerID,
			&rosterID,
			&firstName,
			&lastName,
			&alias,
			&status)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// newly benched player
	benched := store.Player{
		PlayerID:  playerID,
		RosterID:  rosterID,
		FirstName: firstName,
		LastName:  lastName,
		Alias:     alias,
		Status:    status,
	}

	return &store.PlayerChange{
		Active:  active,
		Benched: benched,
	}, tx.Commit()
}
