// Package protocol defines the JSON messages exchanged over /multiplayer.
package protocol

// Message is used for both commands and events. Replies echo the command ID.
// Authentication is the initial, separate {"auth":"x"} message. The server
// replies with "connected" and ClientID, which identifies this connection.
// ClientID and Users identify clients, independently of request/reply ID.
type Message struct {
	Type     string   `json:"type,omitempty"`
	ID       uint64   `json:"id,omitempty"`
	Room     string   `json:"room,omitempty"`
	Rooms    []string `json:"rooms,omitempty"`
	Users    []uint64 `json:"users,omitempty"`
	ClientID uint64   `json:"client_id,omitempty"`
	Text     string   `json:"text,omitempty"`
	Error    string   `json:"error,omitempty"`
}
