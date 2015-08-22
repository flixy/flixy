// Package main provides the executable flixy server
package main

import (
	"encoding/json"
	"fmt"

	"math/rand"
	"net/http"

	log "github.com/Sirupsen/logrus"
	flag "github.com/ogier/pflag"

	"github.com/skyhighwings/flixy/models"

	"github.com/Xe/middleware"
	"github.com/codegangsta/negroni"
	"github.com/drone/routes"
	"github.com/googollee/go-socket.io"
	"github.com/meatballhat/negroni-logrus"
)

// opts is the internal options string.
type options struct {
	Port     int
	Host     string
	LogLevel string
}

var logLevels = map[string]log.Level{
	"panic": log.PanicLevel,
	"fatal": log.FatalLevel,
	"error": log.ErrorLevel,
	"warn":  log.WarnLevel,
	"info":  log.InfoLevel,
	"debug": log.DebugLevel,
}

var (
	opts = options{}
)

// sessions is a map of the session identifier to each Flixy session, generated
// by `makeNewSessionID`
var sessions = make(map[string]*models.Session)

// makeNewSessionID produces a session identifier, which is currently of the
// form "%4d-%4d-%4d-%4d" but this is subject to change and is an
// implementation detail.
func makeNewSessionID() string {
	return fmt.Sprintf("%4d-%4d-%4d-%4d", rand.Intn(9999), rand.Intn(9999), rand.Intn(9999), rand.Intn(9999))
}

// main is the entry point to the flixy server.
func main() {
	flag.IntVarP(&opts.Port, "port", "p", 3000, "the port to listen on")
	flag.StringVarP(&opts.Host, "host", "h", "0.0.0.0", "the host to listen on")
	flag.StringVarP(&opts.LogLevel, "log-level", "l", "info", "the log level to use (possible: panic,fatal,error,warn,info,debug)")

	ll, ok := logLevels[opts.LogLevel]
	if !ok {
		log.Errorf("invalid log level %s set, falling back to default %s", opts.LogLevel, "info")
	}
	log.SetLevel(ll)
	log.Debugf("setting log level to %s", opts.LogLevel)

	log.Info("Starting flixy!")

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// TODO figure out what the fuck is the deal with IDs --- can they be a
	// key in map[session_id]User or something?
	server.On("connection", func(so socketio.Socket) {
		sockid := so.Id()

		so.On("flixy get sync", func(sid string) {
			s, ok := sessions[sid]
			if !ok {
				log.WithFields(log.Fields{
					"verb":          "flixy get sync",
					"member_sockid": sockid,
					"member_remote": so.Request().RemoteAddr,
					"invalid_sid":   sid,
				}).Warn("no video id included")
				so.Emit("flixy invalid session id", sid)
				return
			}

			so.Emit("flixy sync", s.GetWireStatus())
		})

		so.On("flixy new", func(nse map[string]int) {
			log.Infof("client %s creating a new session", sockid)

			sid := makeNewSessionID()

			vid, ok := nse["video_id"]
			if !ok {
				log.WithFields(log.Fields{
					"verb":          "flixy new",
					"member_sockid": sockid,
					"member_remote": so.Request().RemoteAddr,
				}).Warn("no video id included")

				so.Emit("flixy invalid new init map", nse)
				return
			}

			time, ok := nse["time"]
			if !ok {
				log.WithFields(log.Fields{
					"verb":          "flixy new",
					"member_sockid": sockid,
					"member_remote": so.Request().RemoteAddr,
				}).Warn("no time included")

				so.Emit("flixy invalid new init map", nse)
				return
			}

			s := models.NewSession(sid, vid, time)
			s.AddMember(so)
			sessions[sid] = s

			so.Emit("flixy new session", s.ToWireSession())
			log.Infof("new session %s created", sid)
		})

		so.On("flixy pause", func(sid string) {
			log.Infof("%s pausing session %s", sockid, sid)
			s, ok := sessions[sid]
			if !ok {
				log.WithFields(log.Fields{
					"verb":          "flixy pause",
					"member_sockid": sockid,
					"member_remote": so.Request().RemoteAddr,
					"invalid_sid":   sid,
				}).Warn("invalid session id")

				so.Emit("flixy invalid session id", sid)
				return
			}

			s.Pause()
		})

		so.On("flixy play", func(sid string) {
			log.Infof("%s playing session %s", sockid, sid)
			s, ok := sessions[sid]
			if !ok {
				log.WithFields(log.Fields{
					"verb":          "flixy play",
					"member_sockid": sockid,
					"member_remote": so.Request().RemoteAddr,
					"invalid_sid":   sid,
				}).Warn("invalid session id")

				so.Emit("flixy invalid session id", sid)
				return
			}

			s.Play()
			log.Debugf("%s session play being sent to %s", sockid, sid)
		})

		// sid -> session id
		so.On("flixy join", func(sid string) {
			log.Infof("%s joining session %s", sockid, sid)
			s, ok := sessions[sid]
			if !ok {
				log.WithFields(log.Fields{
					"verb":          "flixy join",
					"member_sockid": sockid,
					"member_remote": so.Request().RemoteAddr,
					"invalid_sid":   sid,
				}).Warn("invalid session id")

				so.Emit("flixy invalid session id", sid)
				return
			}

			s.AddMember(so)
		})

		log.Infof("id %s connected", sockid)
	})

	server.On("disconnection", func(so socketio.Socket) {
		sockid := so.Id()

		log.Infof("%v disconnected", sockid)
		log.Debugf("deleting %v's memberships ....", sockid)
		for _, session := range sessions {
			session.RemoveMember(sockid)
		}
	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Error("error:", err)
	})

	mux := http.NewServeMux()
	api := routes.New()

	// TODO remove this
	api.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		// print status here
		// this is just for debugging for now, we need more in-depth stuff soon
		enc := json.NewEncoder(w)
		enc.Encode(sessions)
	})

	// `/sessions/:sid` will 302 the user to the proper Netflix URL if it's a valid SID, setting the session ID in the URL as it does so. It will *also* return the session as a JSON object.
	api.Get("/sessions/:sid", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		session, present := sessions[params.Get(":sid")]
		if !present {
			w.WriteHeader(404)
			return
		}

		w.Header().Set("Location", session.GetNetflixURL())
		w.WriteHeader(302)
		routes.ServeJson(w, session.ToWireSession())
	})

	mux.Handle("/socket.io/", server)
	mux.Handle("/", api)

	n := negroni.New()
	n.Use(negronilogrus.NewMiddleware())
	n.Use(negroni.NewRecovery())
	middleware.Inject(n)
	n.UseHandler(mux)
	n.Run(fmt.Sprintf("%s:%d", opts.Host, opts.Port))
}
