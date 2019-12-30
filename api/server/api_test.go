package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fgrimme/patrongg/store"
	"github.com/fgrimme/patrongg/testdata"
	"github.com/gorilla/mux"
)

type mockRosterStore struct{}

func (rs *mockRosterStore) Get(ctx context.Context, rosterID uint64) (*store.Roster, error) {
	return rosterTests[rosterID].r, rosterTests[rosterID].e
}

// test cases indexed by roster id
var rosterTests = map[uint64]struct {
	d string        // description of test case
	r *store.Roster // mock store response
	e error         // mock store error
	p string        // url path for test requests
	s int           // expected http status code
	b []byte        // expected payload of response
}{
	// url path errors
	0: { // 404
		d: "expect missing id to result in 404",
		p: "roster/",
		s: http.StatusNotFound,
		b: []byte("404 page not found"),
	},
	1: { // 404
		d: "expect invalid status to result in 404",
		p: "roster/1/players/invalid",
		s: http.StatusNotFound,
		b: []byte("404 page not found"),
	},
	// store errors
	2: { // 500
		d: "expect store error to result in 500 when requesting roster",
		p: "roster/2", // roster id is the test case id used by the mock store
		e: errInternal,
		s: http.StatusInternalServerError,
		b: []byte(fmt.Sprintf(`{"error":"%s"}`, errInternal.Error())),
	},
	3: { // 500
		d: "expect store error to result in 500 when requesting active players",
		p: "roster/3/active",
		e: errInternal,
		s: http.StatusInternalServerError,
		b: []byte(fmt.Sprintf(`{"error":"%s"}`, errInternal.Error())),
	},
	// success
	4: { // 200
		d: "expect success when requesting roster",
		p: "roster/4",
		r: testdata.Rosters[382574876546039808].R,
		s: http.StatusOK,
		b: []byte(testdata.Rosters[382574876546039808].RS),
	},
	5: { // 200
		d: "expect success when requesting active players",
		p: "roster/5/active",
		r: testdata.Rosters[382574876546039808].R,
		s: http.StatusOK,
		b: []byte(testdata.Rosters[382574876546039808].AP),
	},
	6: { // 200
		d: "expect success when requesting benched players",
		p: "roster/6/benched",
		r: testdata.Rosters[382574876546039808].R,
		s: http.StatusOK,
		b: []byte(testdata.Rosters[382574876546039808].BP),
	},
}

func TestGet(t *testing.T) {
	// service initialized with a mock store to
	// control the rosters and errors we return
	rs := &rosterService{
		&mockRosterStore{},
		200 * time.Millisecond,
	}

	router := mux.NewRouter()
	router.Handle("/roster/{id:[0-9]+}", rs).Methods("GET")
	router.Handle(fmt.Sprintf("/roster/{id:[0-9]+}/{status:(?:%s|%s)}", Active, Benched), rs).Methods("GET")

	s := httptest.NewServer(router)
	defer s.Close()
	c := s.Client()

	for _, tc := range rosterTests {
		tt := tc
		t.Run(tt.d, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", s.URL, tt.p), nil)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			// expected result
			if want, got := tt.s, resp.StatusCode; want != got {
				t.Errorf("want status code %d got %d", want, got)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			resp.Body.Close()
			if want, got := tt.b, body; string(want) == string(got) {
				t.Errorf("want response\n%+s\ngot\n%+s", want, got)
			}
		})
	}
}

// we use the player id to detemine the return values.
type mockPlayerStore struct{}

func (ps *mockPlayerStore) Insert(ctx context.Context, player store.Player) (*store.Player, error) {
	return insertTests[player.PlayerID].r, insertTests[player.PlayerID].e
}
func (ps *mockPlayerStore) Update(ctx context.Context, player store.Player) (*store.Player, error) {
	return updateTests[player.PlayerID].r, updateTests[player.PlayerID].e
}
func (ps *mockPlayerStore) ChangePlayers(ctx context.Context, players store.PlayerChange) (*store.PlayerChange, error) {
	return nil, nil
}

// test cases indexed by player id
var insertTests = map[uint64]struct {
	d string        // description of test case
	r *store.Player // mock store response
	e error         // mock store error
	u string        // request url path
	p string        // request payload
	s int           // expected http status code
	b []byte        // expected payload
}{
	// url path errors
	0: { // 404
		d: "expect missing path segment to result in 404",
		u: "players/",
		s: http.StatusNotFound,
		b: []byte("404 page not found"),
	},
	// missing body errors
	1: { // 400
		d: "expect missing body to result in 400 when adding player",
		u: "players/add",
		s: http.StatusBadRequest,
		b: []byte(fmt.Sprintf(`{"error":"%s"}`, errBadRequest.Error())),
	},
	// store errors
	2: { // 500
		d: "expect store error to result in 500 when adding player",
		e: errInternal,
		u: "players/add",
		p: `{"player_id":2}`, // id is the testcase-id used by the mock store
		s: http.StatusInternalServerError,
		b: []byte(fmt.Sprintf(`{"error":"%s"}`, errInternal.Error())),
	},
	// success
	3: {
		d: "expect 200 when adding player",
		u: "players/add",
		r: &store.Player{
			PlayerID:  3,
			RosterID:  0,
			FirstName: "foo",
			LastName:  "bar",
			Alias:     "foobar",
			Status:    "active",
		},
		p: `{"player_id":3,"roster_id":0,"first_name":"foo","last_name":"bar","alias":"foobar","active":true}`,
		s: http.StatusOK,
		b: []byte(`{"player_id":3,"roster_id":0,"first_name":"foo","last_name":"bar","alias":"foobar","active":true}`),
	},
}

func TestInsert(t *testing.T) {
	// service initialized with a mock store to
	// control the data and errors we return
	ps := &playerService{
		&mockPlayerStore{},
		200 * time.Millisecond,
	}

	router := mux.NewRouter()
	router.Handle("/players/add", ps).
		Methods("POST")

	s := httptest.NewServer(router)
	defer s.Close()
	c := s.Client()

	for _, tc := range insertTests {
		tt := tc
		t.Run(tt.d, func(t *testing.T) {
			req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", s.URL, tt.u), strings.NewReader(tt.p))
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			// expected result
			if want, got := tt.s, resp.StatusCode; want != got {
				t.Errorf("want status code %d got %d", want, got)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			resp.Body.Close()
			if want, got := tt.b, body; bytes.Compare(want, got) == 1 {
				t.Errorf("want response\n%+s\ngot\n%+s", want, got)
			}
		})
	}
}

// test cases indexed by player id
var updateTests = map[uint64]struct {
	d string        // description of test case
	r *store.Player // mock store response
	e error         // mock store error
	u string        // request url path
	p string        // request payload
	s int           // expected http status code
	b []byte        // expected payload
}{
	// url path errors
	0: { // 500
		d: "expect malformed JSON payload to result in 500 when updating player",
		u: "players/update",
		p: `{"player_id":1`,
		s: http.StatusInternalServerError,
		b: []byte(fmt.Sprintf(`{"error":"%s"}`, errInternal.Error())),
	},
	// store errors
	1: { // 500
		d: "expect store error to result in 500 when updating player",
		e: errInternal,
		u: "players/update",
		p: `{"player_id":1}`,
		s: http.StatusInternalServerError,
		b: []byte(fmt.Sprintf(`{"error":"%s"}`, errInternal.Error())),
	},
	// success
	2: { // 200
		d: "expect player's store to get updated and status get set to benched",
		u: "players/update",
		r: &store.Player{
			PlayerID:  2,
			RosterID:  0,
			FirstName: "foo",
			LastName:  "bar",
			Alias:     "foobar",
			Status:    "benched",
		},
		p: `{"player_id":2,"roster_id":1,"first_name":"foo","last_name":"bar","alias":"foobar","status":"active"}`,
		s: http.StatusOK,
		b: []byte(`{"player_id":2,"roster_id":0,"first_name":"foo","last_name":"bar","alias":"foobar","status":"benched"}`),
	},
}

func TestUpdate(t *testing.T) {
	// service initialized with a mock store to
	// control the data and errors we return
	ps := &playerService{
		&mockPlayerStore{},
		200 * time.Millisecond,
	}

	router := mux.NewRouter()
	router.Handle("/players/update", ps).Methods("PATCH")

	s := httptest.NewServer(router)
	defer s.Close()
	c := s.Client()

	for _, tc := range updateTests {
		tt := tc
		t.Run(tt.d, func(t *testing.T) {
			req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/%s", s.URL, tt.u), strings.NewReader(tt.p))
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			// expected result
			if want, got := tt.s, resp.StatusCode; want != got {
				t.Errorf("want status code %d got %d", want, got)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			resp.Body.Close()
			if want, got := tt.b, body; bytes.Compare(want, got) == 1 {
				t.Errorf("want response\n%+s\ngot\n%+s", want, got)
			}
		})
	}
}
