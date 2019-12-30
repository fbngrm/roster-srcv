package roster

import (
	"context"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fgrimme/patrongg/database"
	"github.com/fgrimme/patrongg/testdata"
)

func TestGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	defer db.Close()

	// before we actually execute our api function, we need to expect required DB actions
	rows := sqlmock.NewRows([]string{"roster_id", "roster_name", "id", "first_name", "last_name", "alias", "active"}).
		AddRow(382574876546039808, "foo", 182919996442279937, "Dominic", "Luklowski", "DataSlayer9", "active").
		AddRow(382574876546039808, "foo", 337332768876789763, "Jane", "Beddingfield", "__Jain", "active").
		AddRow(382574876546039808, "foo", 444322878230495243, "Phillip", "Aaronivic", "phikic", "active").
		AddRow(382574876546039808, "foo", 602403447886839809, "Ji", "Bhok", "TARG3T", "active").
		AddRow(382574876546039808, "foo", 622318474387128331, "Damian", "Grey", "Klikx", "active").
		AddRow(382574876546039808, "foo", 184315303323238400, "Oliver", "Fieldbutter", "Smaayo", "benched")

	query := `SELECT (.+) FROM players as p INNER JOIN rosters ON p.roster_id = rosters.id WHERE p.roster_id = \$1`
	mock.ExpectQuery(query).WillReturnRows(rows)

	want := testdata.Rosters[382574876546039808].R
	rs := New(database.New(db, "mock-db", 0))
	got, err := rs.Get(context.Background(), want.RosterID)
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
