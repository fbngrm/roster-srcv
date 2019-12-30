package store

type Player struct {
	PlayerID  uint64 `json:"player_id"`
	RosterID  uint64 `json:"roster_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Alias     string `json:"alias"`
	Status    string `json:"status"`
}

type PlayerChange struct {
	Active  Player `json:"active"`
	Benched Player `json:"benched"`
}

type Players struct {
	Active  []Player `json:"active"`
	Benched []Player `json:"benched"`
}

type Roster struct {
	RosterID uint64  `json:"roster_id"`
	Name     string  `json:"name"`
	Players  Players `json:"players"`
}
