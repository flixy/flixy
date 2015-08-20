package main

import (
	"time"

	"github.com/googollee/go-socket.io"

	"fmt"
	"net/url"
)

type Session struct {
	SessionID string             `json:"session_id"`
	VideoID   int                `json:"video_id"`
	TrackID   int                `json:"track_id"`
	Time      int                `json:"time"`
	Members   map[string]*Member `json:"members"`
	ticker    *time.Ticker
}

type WireSession struct {
	SessionID string                `json:"session_id"`
	VideoID   int                   `json:"video_id"`
	TrackID   int                   `json:"track_id"`
	Time      int                   `json:"time"`
	Members   map[string]WireMember `json:"members"`
}

type Member struct {
	Socket socketio.Socket
}

type WireMember struct {
}

func (m *Member) SyncTo(time int, vid int, tid int) {
	m.Socket.Emit("flixy sync", map[string]int{"time": time, "video_id": vid, "track_id": tid})
}

func (m *Member) ToWireMember() WireMember {
	// TODO include ID, nick, etc
	return WireMember{}
}

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

func (s *Session) Pause() {
	// for each member, pause
	// also pause the ticker, somehow
}

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

func (s *Session) Sync() {
	for _, member := range s.Members {
		member.SyncTo(s.Time, s.VideoID, s.TrackID)
	}
}

func (s *Session) AddMember(so socketio.Socket) {
	m := &Member{so}
	s.Members[so.Id()] = m
	m.SyncTo(s.Time, s.VideoID, s.TrackID)
}

func (s *Session) RemoveMember(id string) {
	delete(s.Members, id)
}

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
