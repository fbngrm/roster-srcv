package player

import (
	"context"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fgrimme/patrongg/database"
	"github.com/fgrimme/patrongg/store"
)

func TestInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	defer db.Close()

	// before we actually execute our api function, we need to expect required DB actions
	rows := sqlmock.NewRows([]string{"id", "roster_id", "first_name", "last_name", "alias", "active"}).
		AddRow(182919996442279937, 382574876546039808, "Dominic", "Luklowski", "DataSlayer9", "active")

	query := `INSERT INTO players(.*)VALUES(.*) RETURNING *`
	mock.ExpectQuery(query).WillReturnRows(rows)

	want := &store.Player{
		PlayerID:  182919996442279937,
		RosterID:  382574876546039808,
		FirstName: "Dominic",
		LastName:  "Luklowski",
		Alias:     "DataSlayer9",
		Status:    "active",
	}
	ps := New(database.New(db, "mock-db", 0))
	got, err := ps.Insert(context.Background(), *want)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want\n%+v\ngot\n%+v", want, got)
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// changes the roster id and status
func TestUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	defer db.Close()

	// before we actually execute our api function, we need to expect required DB actions
	rows := sqlmock.NewRows([]string{"id", "roster_id", "first_name", "last_name", "alias", "active"}).
		AddRow(182919996442279937, 1, "Dominic", "Luklowski", "DataSlayer9", "benched")

	query := `
  UPDATE players
  SET
    roster_id = COALESCE\(NULLIF\(\$2, CAST\(0 AS BIGINT\)\), players.roster_id\),
    first_name = COALESCE\(NULLIF\(\$3, \'\'\), players.first_name\),
    last_name  = COALESCE\(NULLIF\(\$4, \'\'\), players.last_name\),
    alias      = COALESCE\(NULLIF\(\$5, \'\'\), players.alias\),
	status     = COALESCE\(NULLIF\(\$6, \'\'\), players.status\)
  WHERE id = \$1
  RETURNING \*`

	p := store.Player{
		PlayerID:  182919996442279937,
		RosterID:  382574876546039808,
		FirstName: "Dominic",
		LastName:  "Luklowski",
		Alias:     "DataSlayer9",
		Status:    "active",
	}

	mock.ExpectQuery(query).WithArgs(
		p.PlayerID,
		p.RosterID,
		p.FirstName,
		p.LastName,
		p.Alias,
		p.Status,
	).WillReturnRows(rows)

	ps := New(database.New(db, "mock-db", 0))
	got, err := ps.Update(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	want := p
	want.RosterID = 1
	want.Status = "benched"
	if !reflect.DeepEqual(got, &want) {
		t.Errorf("want\n%+v\ngot\n%+v\n", want, got)
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// swaps the status
func TestChangePlayer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	defer db.Close()

	// before we actually execute our api function, we need to expect required DB actions
	rowsActive := sqlmock.NewRows([]string{"id", "roster_id", "first_name", "last_name", "alias", "active"}).
		AddRow(2, 1, "boo", "baz", "boobaz", "active")
	rowsBenched := sqlmock.NewRows([]string{"id", "roster_id", "first_name", "last_name", "alias", "active"}).
		AddRow(1, 1, "foo", "bar", "foobar", "benched")

	query := `
  UPDATE players
  SET status = \$1
  WHERE id = \$2
  AND status = \$3
  AND roster_id = \( \-\- ensures that both players are in the same roster
    SELECT roster_id
    FROM players
    WHERE id = \$4
  \)
  RETURNING *`

	mock.ExpectBegin()
	mock.ExpectPrepare(query)

	// expected query for benching active player
	mock.ExpectQuery(query).WithArgs(
		"active",
		2,
		"benched",
		1,
	).WillReturnRows(rowsActive)

	// expected query for activating benched player
	mock.ExpectQuery(query).WithArgs(
		"benched",
		1,
		"active",
		2,
	).WillReturnRows(rowsBenched)

	mock.ExpectCommit()

	// input
	players := store.PlayerChange{
		Active: store.Player{
			PlayerID:  1,
			RosterID:  1,
			FirstName: "foo",
			LastName:  "bar",
			Alias:     "foobar",
			Status:    "active",
		},
		Benched: store.Player{
			PlayerID:  2,
			RosterID:  1,
			FirstName: "boo",
			LastName:  "baz",
			Alias:     "boobaz",
			Status:    "benched",
		},
	}

	// expected result
	want := &store.PlayerChange{
		Active: store.Player{
			PlayerID:  2,
			RosterID:  1,
			FirstName: "boo",
			LastName:  "baz",
			Alias:     "boobaz",
			Status:    "active",
		},
		Benched: store.Player{
			PlayerID:  1,
			RosterID:  1,
			FirstName: "foo",
			LastName:  "bar",
			Alias:     "foobar",
			Status:    "benched",
		},
	}

	ps := New(database.New(db, "mock-db", 0))
	got, err := ps.ChangePlayers(context.Background(), players)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want\n%+v\ngot\n%+v\n", want, got)
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
