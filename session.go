package main

import (
	"github.com/googollee/go-socket.io"
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
