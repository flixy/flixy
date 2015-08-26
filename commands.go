package main

import (
	"encoding/json"

	log "github.com/flixy/flixy/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/flixy/flixy/Godeps/_workspace/src/github.com/googollee/go-socket.io"
	"github.com/flixy/flixy/models"
)

// SyncHandler returns the handler for `flixy get sync`.
func SyncHandler(so socketio.Socket) func(string) {
	sockid := so.Id()
	sockip := so.Request().RemoteAddr

	return func(jsonmsg string) {
		var data models.GetSyncMessage

		if err := json.Unmarshal([]byte(jsonmsg), &data); err != nil {
			log.WithFields(log.Fields{
				"verb":          "flixy get sync",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Error(err)
			return
		}

		sid := data.SessionID

		s, ok := sessions[sid]
		if !ok {
			log.WithFields(log.Fields{
				"verb":          "flixy get sync",
				"member_sockid": sockid,
				"member_remote": sockip,
				"invalid_sid":   sid,
			}).Warn("invalid session id")
			so.Emit("flixy invalid session id", sid)
			return
		}

		log.WithFields(log.Fields{
			"verb":          "flixy get sync",
			"member_sockid": sockid,
			"member_remote": sockip,
		}).Debug("getting sync state")
		so.Emit("flixy sync", s.GetWireSession())
	}
}

// NewHandler returns the handler for `flixy new`.
func NewHandler(so socketio.Socket) func(string) {
	sockid := so.Id()
	sockip := so.Request().RemoteAddr

	return func(jsonmsg string) {
		var data models.NewMessage

		if err := json.Unmarshal([]byte(jsonmsg), &data); err != nil {
			log.WithFields(log.Fields{
				"verb":          "flixy new",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Error(err)
		}

		log.Infof("client %s creating a new session", sockid)

		sid := makeNewSessionID()

		vid := data.VideoID
		if vid == 0 {
			log.WithFields(log.Fields{
				"verb":          "flixy new",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Warn("invalid video id?")

			so.Emit("flixy invalid new data", jsonmsg)
			return
		}

		time := data.Time
		nick := data.Nick
		if nick == "" {
			nick = "(no nick)"
		}

		s := models.NewSession(sid, vid, time)
		m := s.AddMember(so, nick)
		members[sockid] = m
		sessions[sid] = s

		so.Emit("flixy new session", s.GetWireSession())
		// TODO make this the new style logging :-)
		log.Infof("new session %s created", sid)
	}
}

// PauseHandler returns the handler for `flixy pause`.
func PauseHandler(so socketio.Socket) func(string) {
	sockid := so.Id()
	sockip := so.Request().RemoteAddr

	return func(jsonmsg string) {
		var data models.PauseMessage

		if err := json.Unmarshal([]byte(jsonmsg), &data); err != nil {
			log.WithFields(log.Fields{
				"verb":          "flixy pause",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Error(err)
			return
		}

		sid := data.SessionID

		log.Infof("%s pausing session %s", sockid, sid)
		s, ok := sessions[sid]
		if !ok {
			log.WithFields(log.Fields{
				"verb":          "flixy pause",
				"member_sockid": sockid,
				"member_remote": sockip,
				"invalid_sid":   sid,
			}).Warn("invalid session id")

			so.Emit("flixy invalid session id", sid)
			return
		}

		s.Pause()
		log.WithFields(log.Fields{
			"verb":          "flixy pause",
			"member_sockid": sockid,
			"member_remote": sockip,
		}).Debug("pausing")
	}
}

// PlayHandler returns the handler for `flixy play`.
func PlayHandler(so socketio.Socket) func(string) {
	sockid := so.Id()
	sockip := so.Request().RemoteAddr

	return func(jsonmsg string) {
		var data models.PlayMessage

		if err := json.Unmarshal([]byte(jsonmsg), &data); err != nil {
			log.WithFields(log.Fields{
				"verb":          "flixy play",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Error(err)
			return
		}

		sid := data.SessionID

		log.Infof("%s playing session %s", sockid, sid)
		s, ok := sessions[sid]
		if !ok {
			log.WithFields(log.Fields{
				"verb":          "flixy play",
				"member_sockid": sockid,
				"member_remote": sockip,
				"invalid_sid":   sid,
			}).Warn("invalid session id")

			so.Emit("flixy invalid session id", sid)
			return
		}

		s.Play()
		log.WithFields(log.Fields{
			"verb":          "flixy play",
			"member_sockid": sockid,
			"member_remote": sockip,
		}).Debug("playing")
	}
}

// JoinHandler returns the handler for `flixy join`.
func JoinHandler(so socketio.Socket) func(string) {
	sockid := so.Id()
	sockip := so.Request().RemoteAddr

	return func(jsonmsg string) {
		var data models.JoinMessage

		if err := json.Unmarshal([]byte(jsonmsg), &data); err != nil {
			log.WithFields(log.Fields{
				"verb":          "flixy join",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Error(err)
			return
		}

		sid := data.SessionID
		nick := data.Nick
		if nick == "" {
			nick = "(no nick)"
		}

		s, ok := sessions[sid]
		if !ok {
			log.WithFields(log.Fields{
				"verb":          "flixy join",
				"member_sockid": sockid,
				"member_remote": sockip,
				"invalid_sid":   sid,
			}).Warn("invalid session id")

			so.Emit("flixy invalid session id", sid)
			return
		}

		s.AddMember(so, nick)

		log.WithFields(log.Fields{
			"verb":          "flixy play",
			"member_sockid": sockid,
			"member_remote": sockip,
			"session_id":    sid,
		}).Debug("joining a session")
	}
}

// SeekHandler returns the handler for `flixy seek`.
func SeekHandler(so socketio.Socket) func(string) {
	sockid := so.Id()
	sockip := so.Request().RemoteAddr

	return func(jsonmsg string) {
		var data models.SeekMessage

		if err := json.Unmarshal([]byte(jsonmsg), &data); err != nil {
			log.WithFields(log.Fields{
				"verb":          "flixy seek",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Error(err)
			return
		}

		sid := data.SessionID
		ts := data.Time

		s, ok := sessions[sid]
		if !ok {
			log.WithFields(log.Fields{
				"verb":          "flixy seek",
				"member_sockid": sockid,
				"member_remote": sockip,
				"invalid_sid":   sid,
			}).Warn("invalid session id")
			so.Emit("flixy invalid session id", sid)
			return
		}

		log.WithFields(log.Fields{
			"verb":          "flixy seek",
			"member_sockid": sockid,
			"member_remote": sockip,
			"session":       sid,
		}).Debug("setting time")
		s.SetTime(ts)
	}
}
