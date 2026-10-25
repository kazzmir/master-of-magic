package protocol

// GameState is an authoritative game setup snapshot. IDs, not display names,
// identify games and their players. Started means both players are ready; each
// participant then receives a private combat-start message with a relay token.
type GameState struct {
	ID         uint64       `json:"id"`
	Room       string       `json:"room"`
	HostID     uint64       `json:"host_id"`
	Players    []GamePlayer `json:"players"`
	Started    bool         `json:"started"`
	Budget     uint64       `json:"budget"`
	AttackerID uint64       `json:"attacker_id"`
	Revision   uint64       `json:"revision"`
}

func (state GameState) Clone() GameState {
	players := make([]GamePlayer, len(state.Players))
	copy(players, state.Players)
	for i := range players {
		players[i].Loadout = players[i].Loadout.Clone()
	}
	state.Players = players
	return state
}

type GamePlayer struct {
	ClientID uint64  `json:"client_id"`
	Name     string  `json:"name"`
	Ready    bool    `json:"ready"`
	Loadout  Loadout `json:"loadout"`
}
