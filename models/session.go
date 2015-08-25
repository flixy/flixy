package models

import (
	"fmt"
	"net/url"
	"time"

	"github.com/e-dard/tock"
	"github.com/googollee/go-socket.io"
)

// Session is the *internal* representation of a flixy session, which is a
// collection of *Members*, along with:
//   - A single Session ID, which is the name by which this is referred (this is probably always going to be the key in the `sessions` map in `main.go`
//   - A single Video ID (a session can only be watching one thing at a time)
//   - A single Time, which is a JS time in milliseconds (recorded as an `int`, but is that enough for a JS timestamp (which is milliseconds)?) Possible TODO is to make this a Time in the internal rep.
//   - A bool indicating whether or not the session is currently paused
type Session struct {
	SessionID string             `json:"session_id"`
	VideoID   int                `json:"video_id"`
	Time      int                `json:"time"`
	Members   map[string]*Member `json:"members"`
	Paused    bool               `json:"paused"`
	ticker    *tock.Ticker
}

// WireSession is the *external* representation of a flixy session. It has no
// references to anything that has an unexported field, as that currently
// (2015-08-20) causes reflection errors.
// It is comprised of the session ID, the video ID, the time, whether or not
// the session is paused, and the members.
type WireSession struct {
	SessionID string                `json:"session_id"`
	VideoID   int                   `json:"video_id"`
	Time      int                   `json:"time"`
	Paused    bool                  `json:"paused"`
	Members   map[string]WireMember `json:"members"`
}

// NewSession creates and return a new `Session` with the given arguments,
// starting paused.
func NewSession(id string, vid int, ts int) *Session {
	// TODO add an option to start unpaused?
	s := Session{
		id,
		vid,
		ts,
		make(map[string]*Member),
		true,
		tock.NewTicker(time.Millisecond),
	}
	s.ticker.Stop()

	// TODO run a select{} on this for a quit channel for when we delete
	// the session, so we don't continue accumulating goroutines for no
	// reason over the life of the application

	go func() {
		for {
			<-s.ticker.C
			s.Time++
		}
	}()

	return &s
}

func (s *Session) SetTime(ts int) {
	s.Time = ts
	s.Sync()
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

// GetWireSession returns a `WireSession` from a given `Session`, which is a
// sanitized version of a session suitable for sending over a socket.io
// connection.
func (s *Session) GetWireSession() WireSession {
	wms := make(map[string]WireMember)
	for k, member := range s.Members {
		wms[k] = member.ToWireMember()
	}
	return WireSession{
		s.SessionID,
		s.VideoID,
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
func (s *Session) AddMember(so socketio.Socket, nick string) *Member {
	m := &Member{so, s, nick}
	s.Members[so.Id()] = m
	m.Sync()

	// Touching the member's socket directly feels wrong. This should
	// probably become non-exported.
	m.Socket.Emit("flixy join session", s.GetWireSession())

	return m
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
	params.Add("flixySessionId", s.SessionID)
	u.RawQuery = params.Encode()

	return u.String()
}
