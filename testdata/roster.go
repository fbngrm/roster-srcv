package testdata

import "github.com/fgrimme/patrongg/store"

var Rosters = map[uint64]struct {
	R  *store.Roster // roster
	RS string        // JSON represetation of entire roster
	AP string        // JSON represetation of active players
	BP string        // JSON represetation of benched players

}{
	382574876546039808: {
		R: &store.Roster{
			RosterID: 382574876546039808,
			Name:     "foo",
			Players: store.Players{
				Active: []store.Player{
					{
						PlayerID:  182919996442279937,
						RosterID:  382574876546039808,
						FirstName: "Dominic",
						LastName:  "Luklowski",
						Alias:     "DataSlayer9",
						Status:    "active",
					},
					{
						PlayerID:  337332768876789763,
						RosterID:  382574876546039808,
						FirstName: "Jane",
						LastName:  "Beddingfield",
						Alias:     "__Jain",
						Status:    "active",
					},
					{
						PlayerID:  444322878230495243,
						RosterID:  382574876546039808,
						FirstName: "Phillip",
						LastName:  "Aaronivic",
						Alias:     "phikic",
						Status:    "active",
					},
					{
						PlayerID:  602403447886839809,
						RosterID:  382574876546039808,
						FirstName: "Ji",
						LastName:  "Bhok",
						Alias:     "TARG3T",
						Status:    "active",
					},
					{
						PlayerID:  622318474387128331,
						RosterID:  382574876546039808,
						FirstName: "Damian",
						LastName:  "Grey",
						Alias:     "Klikx",
						Status:    "active",
					},
				},
				Benched: []store.Player{
					{
						PlayerID:  184315303323238400,
						RosterID:  382574876546039808,
						FirstName: "Oliver",
						LastName:  "Fieldbutter",
						Alias:     "Smaayo",
						Status:    "benched",
					},
				},
			},
		},
		RS: `{"id":382574876546039808,"name":"foo","players":{"status":[{"id":182919996442279937,"roster_id":382574876546039808,"first_name":"Dominic","last_name":"Luklowski","alias":"DataSlayer9","status":"active"},{"id":337332768876789763,"roster_id":382574876546039808,"first_name":"Jane","last_name":"Beddingfield","alias":"__Jain","status":"active"},{"id":444322878230495243,"roster_id":382574876546039808,"first_name":"Phillip","last_name":"Aaronivic","alias":"phikic","status":"active"},{"id":602403447886839809,"roster_id":382574876546039808,"first_name":"Ji","last_name":"Bhok","alias":"TARG3T","status":"active"},{"id":622318474387128331,"roster_id":382574876546039808,"first_name":"Damian","last_name":"Grey","alias":"Klikx","status":"active"}],"benched":[{"id":184315303323238400,"roster_id":382574876546039808,"first_name":"Oliver","last_name":"Fieldbutter","alias":"Smaayo","status":"benched"}]}}`,
		AP: `[{"id":182919996442279937,"roster_id":382574876546039808,"first_name":"Dominic","last_name":"Luklowski","alias":"DataSlayer9","status":"active"},{"id":337332768876789763,"roster_id":382574876546039808,"first_name":"Jane","last_name":"Beddingfield","alias":"__Jain","status":"active"},{"id":444322878230495243,"roster_id":382574876546039808,"first_name":"Phillip","last_name":"Aaronivic","alias":"phikic","status":"active"},{"id":602403447886839809,"roster_id":382574876546039808,"first_name":"Ji","last_name":"Bhok","alias":"TARG3T","status":"active"},{"id":622318474387128331,"roster_id":382574876546039808,"first_name":"Damian","last_name":"Grey","alias":"Klikx","status":"active"}]`,
		BP: `[{"id":184315303323238400,"roster_id":382574876546039808,"first_name":"Oliver","last_name":"Fieldbutter","alias":"Smaayo","status":"benched"}]`,
	},
}
