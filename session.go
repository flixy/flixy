package main

import (
	"github.com/googollee/go-socket.io"

	"fmt"
	"net/url"
)

type Session struct {
	VideoID int                `json:"video_id"`
	TrackID int                `json:"track_id"`
	Time    int                `json:"time"`
	Members map[string]*Member `json:"members"`
}

type Member struct {
	Socket socketio.Socket
}

func (m *Member) SyncTo(time int, vid int, tid int) {
	m.Socket.Emit("flixy sync", map[string]int{"time": time, "video_id": vid, "track_id": tid})
}

func (s *Session) Sync() {
	for _, member := range s.Members {
		member.SyncTo(s.Time, s.VideoID, s.TrackID)
	}
}

func (s *Session) AddMember(so socketio.Socket) {
	s.Members[so.Id()] = &Member{so}
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
	u.RawQuery = params.Encode()

	return u.String()
}
