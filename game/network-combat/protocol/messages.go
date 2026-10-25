// Package protocol defines the JSON messages exchanged over /multiplayer.
package protocol

// Message is used for both commands and events. Replies echo the command ID.
// Authentication is the initial, separate {"auth":"x"} message.
type Message struct {
	Type  string   `json:"type,omitempty"`
	ID    uint64   `json:"id,omitempty"`
	Room  string   `json:"room,omitempty"`
	Rooms []string `json:"rooms,omitempty"`
	Users []string `json:"users,omitempty"`
	User  string   `json:"user,omitempty"`
	Text  string   `json:"text,omitempty"`
	Error string   `json:"error,omitempty"`
}
