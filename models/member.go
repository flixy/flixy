package models

import "github.com/flixy/flixy/Godeps/_workspace/src/github.com/googollee/go-socket.io"

// Member is the *internal* representation of a member of a flixy session. It
// currently has only a socket, but will have a `nickname` or something like it
// in the near future.
type Member struct {
	Socket socketio.Socket
	*Session
	Nick string `json:"nick"`
}

// WireMember is the *external* representation of a member of a flixy session.
// It has nothing in it currently, but will have a `nickname` or something like
// it in the neat future.
type WireMember struct {
	Nick string `json:"nick"`
}

// Sync tells the given member the state of the session.
func (m *Member) Sync() {
	m.Socket.Emit("flixy sync", m.Session.GetWireSession())
}

// ToWireMember converts a given `Member` to a `WireMember`, which
// sanitizes the arguments to be suitable to be sent over a socket.io
// connection.
func (m *Member) ToWireMember() WireMember {
	// TODO include ID, nick, etc
	return WireMember{m.Nick}
}

// LeaveSession tells the member's session to remove the member's socket ID,
// thus removing the member from the session and any further sync, etc updates.
func (m *Member) LeaveSession() {
	m.Session.RemoveMember(m.Socket.Id())
}
