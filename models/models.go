// Package models provides the models and methods to interface with them for
// the flixy server.
package models

import (
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
type Session struct {
	SessionID string             `json:"session_id"`
	VideoID   int                `json:"video_id"`
	TrackID   int                `json:"track_id"`
	Time      int                `json:"time"`
	Members   map[string]*Member `json:"members"`
	ticker    *time.Ticker
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
	Members   map[string]WireMember `json:"members"`
}

// Member is the *internal* representation of a member of a flixy session. It
// currently has only a socket, but will have a `nickname` or something like it
// in the near future.
type Member struct {
	Socket socketio.Socket
}

// WireMember is the *external* representation of a member of a flixy session.
// It has nothing in it currently, but will have a `nickname` or something like
// it in the neat future.
type WireMember struct {
}

// SyncTo sends a message over the wire to the given `Member` to inform
// their client to sync the video state to the arguments.
func (m *Member) SyncTo(time int, vid int, tid int) {
	m.Socket.Emit("flixy sync", map[string]int{"time": time, "video_id": vid, "track_id": tid})
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
		id, vid, tid, ts, make(map[string]*Member), time.NewTicker(time.Millisecond),
	}

	go func() {
		for {
			<-s.ticker.C
			s.Time++
		}
	}()

	return &s
}

// Pause pauses the server-side ticker of a given `Session` and inform all
// clients that they should be paused, too.
func (s *Session) Pause() {
	// for each member, pause
	// also pause the ticker, somehow
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
		wms,
	}
}

// Sync syncs all members of a given session to the session's idea of where
// everyone should be.
func (s *Session) Sync() {
	for _, member := range s.Members {
		member.SyncTo(s.Time, s.VideoID, s.TrackID)
	}
}

// AddMember adds a member to the given session and syncs them to where the
// server is.
func (s *Session) AddMember(so socketio.Socket) {
	m := &Member{so}
	s.Members[so.Id()] = m
	m.SyncTo(s.Time, s.VideoID, s.TrackID)
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
