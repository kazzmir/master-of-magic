package protocol

// GameState is an authoritative game setup snapshot. IDs, not display names,
// identify games and their players. Started currently marks a future start hook.
type GameState struct {
	ID      uint64       `json:"id"`
	Room    string       `json:"room"`
	HostID  uint64       `json:"host_id"`
	Players []GamePlayer `json:"players"`
	Started bool         `json:"started"`
}

type GamePlayer struct {
	ClientID uint64 `json:"client_id"`
	Name     string `json:"name"`
	Ready    bool   `json:"ready"`
}
