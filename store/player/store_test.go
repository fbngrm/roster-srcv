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

// func TestUpdate(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("unexpected error %v", err)
// 	}
// 	defer db.Close()

// 	query := `UPDATE players SET roster_id = \$1 WHERE id = \$2`
// 	mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(1, 1))

// 	p := store.Player{
// 		PlayerID:  182919996442279937,
// 		RosterID:  382574876546039808,
// 		FirstName: "Dominic",
// 		LastName:  "Luklowski",
// 		Alias:     "DataSlayer9",
// 		Status:    "active",
// 	}
// 	ps := New(database.New(db, "mock-db", 0))
// 	got, err := ps.Update(context.Background(), p)
// 	if err != nil {
// 		t.Fatalf("unexpected error %v", err)
// 	}
// 	if got != 1 {
// 		t.Errorf("want 1 got %d", got)
// 	}
// 	// we make sure that all expectations were met
// 	if err := mock.ExpectationsWereMet(); err != nil {
// 		t.Errorf("there were unfulfilled expectations: %s", err)
// 	}
// }
