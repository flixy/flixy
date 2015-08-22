// Package models provides the models and methods to interface with them for
// the flixy server.
package models

import (
	"github.com/e-dard/tock"
	"time"

	"github.com/googollee/go-socket.io"

	"fmt"
	"net/url"
)

// Session is the *internal* representation of a flixy session, which is a
// collection of *Members*, along with:
//   - A single Session ID, which is the name by which this is referred (this is probably always going to be the key in the `sessions` map in `main.go`
//   - a single Video ID (a session can only be watching one thing at a time)
//   - a single Track ID (it's unknown what this is for, it's potentially a tracking ID so we should possibly not include it.
//   - A single Time, which is a JS time in milliseconds (recorded as an `int`, but is that enough for a JS timestamp (which is milliseconds)?) Possible TODO is to make this a Time in the internal rep.
//   - a stop channel, which listens for messages and stops the ticker. For internal use only.
type Session struct {
	SessionID string             `json:"session_id"`
	VideoID   int                `json:"video_id"`
	TrackID   int                `json:"track_id"`
	Time      int                `json:"time"`
	Members   map[string]*Member `json:"members"`
	Paused    bool               `json:"paused"`
	ticker    *tock.Ticker
}

// WireSession is the *external* representation of a flixy session. It has no
// references to anything that has an unexported field, as that currently
// (2015-08-20) causes reflection errors.
// It is comprised of the session ID, the video ID, the track ID, the time, and
// the members.
type WireSession struct {
	SessionID string                `json:"session_id"`
	VideoID   int                   `json:"video_id"`
	TrackID   int                   `json:"track_id"`
	Time      int                   `json:"time"`
	Paused    bool                  `json:"paused"`
	Members   map[string]WireMember `json:"members"`
}

// Member is the *internal* representation of a member of a flixy session. It
// currently has only a socket, but will have a `nickname` or something like it
// in the near future.
type Member struct {
	Socket socketio.Socket
	*Session
}

// WireMember is the *external* representation of a member of a flixy session.
// It has nothing in it currently, but will have a `nickname` or something like
// it in the neat future.
type WireMember struct {
}

// WireStatus is the *external* representation of the current status of a flixy
// session, suitable for being sent over a go-socket.io connection.
type WireStatus struct {
	VideoID int  `json:"video_id"`
	TrackID int  `json:"track_id"`
	Time    int  `json:"time"`
	Paused  bool `json:"paused"`
}

// Sync tells the given member the state of the session.
func (m *Member) Sync() {
	m.Socket.Emit("flixy sync", m.Session.GetWireStatus())
}

// ToWireMember converts a given `Member` to a `WireMember`, which
// sanitizes the arguments to be suitable to be sent over a socket.io
// connection.
func (m *Member) ToWireMember() WireMember {
	// TODO include ID, nick, etc
	return WireMember{}
}

// NewSession creates and return a new `Session` with the given arguments,
// starting the ticker for it in the process.
func NewSession(id string, vid int, tid int, ts int) *Session {
	s := Session{
		id,
		vid,
		tid,
		ts,
		make(map[string]*Member),
		false,
		tock.NewTicker(time.Millisecond),
	}

	go func() {
		for {
			<-s.ticker.C
			s.Time++
		}
	}()

	return &s
}

// GetWireStatus returns a wire representation of where the session is without
// including the session ID or members.
func (s *Session) GetWireStatus() WireStatus {
	ws := WireStatus{
		s.VideoID,
		s.TrackID,
		s.Time,
		s.Paused,
	}

	return ws
}

// SendToAll emits a given eventName on all member sockets, with the given
// message. Please don't pass anything that has an unexported struct key
// anywhere in it at all to this. go-socket.io will choke on it.
func (s *Session) SendToAll(eventName string, message interface{}) {
	for _, m := range s.Members {
		m.Socket.Emit(eventName, message)
	}
}

// Play starts the server-side ticker of a given Session and informs all
// Members that it is time to resume playing again.
func (s *Session) Play() {
	s.ticker.Resume()
	s.Paused = false

	// TODO should this have its own dedicated `flixy play` event?
	s.Sync()
}

// Pause pauses the server-side ticker of a given `Session` and inform all
// clients that they should be paused, too.
func (s *Session) Pause() {
	s.ticker.Stop()
	s.Paused = true

	// TODO should this have its own dedicated `flixy pause` event?
	s.Sync()
}

// ToWireSession returns a `WireSession` from a given `Session`, which is a
// sanitized version of a session suitable for sending over a socket.io
// connection.
func (s *Session) ToWireSession() WireSession {
	wms := make(map[string]WireMember)
	for k, member := range s.Members {
		wms[k] = member.ToWireMember()
	}
	return WireSession{
		s.SessionID,
		s.VideoID,
		s.TrackID,
		s.Time,
		s.Paused,
		wms,
	}
}

// Sync syncs all members of a given session to the session's idea of where
// everyone should be.
func (s *Session) Sync() {
	// TODO should this be using `s.SendToAll` instead?
	for _, member := range s.Members {
		member.Sync()
	}
}

// AddMember adds a member to the given session and syncs them to where the
// server is.
func (s *Session) AddMember(so socketio.Socket) {
	m := &Member{so, s}
	s.Members[so.Id()] = m
	m.Sync()

	// Touching the member's socket directly feels wrong. This should
	// probably become non-exported.
	m.Socket.Emit("flixy join session", s.ToWireSession())
}

// RemoveMember removes a member from the given session.
func (s *Session) RemoveMember(id string) {
	delete(s.Members, id)
}

// GetNetflixURL returns the Netflix URL to which a user should be redirected
// to so that they will be on the same video as the server initially.
func (s *Session) GetNetflixURL() string {
	// is this totally overengineered? should I just string cat these
	// together?
	var u *url.URL
	u, err := url.Parse("https://www.netflix.com")
	if err != nil {
		// something has gone *seriously* wrong
		panic("URL parsing a simple URL failed")
	}

	u.Path += fmt.Sprintf("/watch/%d", s.VideoID)
	params := url.Values{}
	params.Add("trackId", fmt.Sprintf("%d", s.TrackID))
	params.Add("flixySessionId", s.SessionID)
	u.RawQuery = params.Encode()

	return u.String()
}