package roster

import (
	"context"

	"github.com/fgrimme/patrongg/database"
	"github.com/fgrimme/patrongg/store"
)

// RosterStore handles operations on the the
// rosters table of the encapsulated datastore.
type RosterStore struct {
	db *database.DB
}

func New(db *database.DB) *RosterStore {
	return &RosterStore{
		db: db,
	}
}

// Get returns a representation of the entire roster for the given id or an error.
func (rs *RosterStore) Get(ctx context.Context, rosterID uint64) (*store.Roster, error) {
	query := `
  SELECT
    rosters.id,
    rosters.name,
    p.id,
    p.first_name,
    p.last_name,
    p.alias,
    p.status
  FROM players as p
  INNER JOIN rosters ON p.roster_id = rosters.id
  WHERE p.roster_id = $1`

	db := rs.db.GetDB()
	ctx, cancel := rs.db.RequestContext(ctx)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, rosterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := store.Players{
		Active:  make([]store.Player, 0),
		Benched: make([]store.Player, 0),
	}
	var id uint64
	var rosterName string
	var playerID uint64
	var firstName string
	var lastName string
	var alias string
	var status string
	for rows.Next() {
		if err := rows.Scan(
			&id,
			&rosterName,
			&playerID,
			&firstName,
			&lastName,
			&alias,
			&status,
		); err != nil {
			return nil, err
		}
		p := store.Player{
			PlayerID:  playerID,
			RosterID:  id,
			FirstName: firstName,
			LastName:  lastName,
			Alias:     alias,
			Status:    status,
		}
		if status == "active" {
			players.Active = append(players.Active, p)
		} else {
			players.Benched = append(players.Benched, p)
		}
	}
	return &store.Roster{
		RosterID: id,
		Name:     rosterName,
		Players:  players,
	}, rows.Err()
}
