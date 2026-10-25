// Package protocol defines the JSON messages exchanged over /multiplayer.
package protocol

const MaxRoomNameLength = 30
const MaxUserNameLength = 30

const ClientVersion = 1
const ServerVersion = 1

// Message is used for both commands and events. Replies echo the command ID.
// Authentication is the initial, separate {"auth":"x","name":"Player1234"} message. The server
// replies with "connected" and ClientID, which identifies this connection.
// Both sides send a "version" message after authentication succeeds; versions
// are informational for now. ClientID and Users identify clients, independently
// of request/reply ID.
type Message struct {
	Type     string   `json:"type,omitempty"`
	ID       uint64   `json:"id,omitempty"`
	Version  int      `json:"version,omitempty"`
	Room     string   `json:"room,omitempty"`
	Private  bool     `json:"private,omitempty"`
	Password string   `json:"password,omitempty"`
	Rooms    []string `json:"rooms,omitempty"`
	Users    []uint64 `json:"users,omitempty"`
	ClientID uint64   `json:"client_id,omitempty"`
	Name     string   `json:"name,omitempty"`
	// Names accompanies room-users snapshots; keys are client IDs, not names.
	Names map[uint64]string `json:"names,omitempty"`
	Text  string            `json:"text,omitempty"`
	Error string            `json:"error,omitempty"`
}
