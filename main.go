// Package main provides the executable flixy server
package main

import (
	"encoding/json"
	"fmt"

	"log"
	"math/rand"
	"net/http"

	"github.com/skyhighwings/flixy/models"

	"github.com/Xe/middleware"
	"github.com/codegangsta/negroni"
	"github.com/drone/routes"
	"github.com/googollee/go-socket.io"
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
	fmt.Println("Hello, world!")

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	sessions["test"] = &models.Session{SessionID: "test", VideoID: 1, TrackID: 2, Time: 3, Members: make(map[string]*models.Member)}

	// TODO figure out what the fuck is the deal with IDs --- can they be a
	// key in map[session_id]User or something?
	server.On("connection", func(so socketio.Socket) {
		s := sessions["test"]
		log.Printf("34 %v", s)

		// se -> sync event
		// map[string]interface{} that looks like
		// { paused: false, ts: 128519241124, vid: 6527312, tid: 5124623 }
		so.On("flixy sync", func(se map[string]interface{}) {
			log.Printf("%v", se)
		})

		/*
			`flixy new` creates a new session, and accepts a map of the form:
			{
				video_id: 6135412,
				track_id: 51251265,
				time: 12423552346
			}

			upon registration it will emit an event `flixy new session` to the client, which will be a map:
			{
				session_id: "1254-5231-5432-4324",
				video_id: 1523543,
				track_id: 236523,
				time: 523623662,
				members: {}
			}

			TODO can this `members` be reasonably removed?

		*/
		so.On("flixy new", func(nse map[string]int) {
			sid := makeNewSessionID()

			vid, ok := nse["video_id"]
			if !ok {
				log.Fatalf("`flixy new` from %s (%s) had no video_id", so.Id(), so.Request().RemoteAddr)
			}

			tid, ok := nse["track_id"]
			if !ok {
				log.Fatalf("`flixy new` from %s (%s) had no track_id", so.Id(), so.Request().RemoteAddr)
			}

			time, ok := nse["time"]
			if !ok {
				log.Fatalf("`flixy new` from %s (%s) had no time", so.Id(), so.Request().RemoteAddr)
			}

			s := models.NewSession(sid, vid, tid, time)
			s.AddMember(so)
			sessions[sid] = s

			so.Emit("flixy new session", s.ToWireSession())
		})

		so.On("flixy pause", func(sid string) {
			s, ok := sessions[sid]
			if !ok {
				log.Fatalf("`flixy pause` from %s (%s) had an invalid session_id", so.Id(), so.Request().RemoteAddr)
			}

			s.Pause()
		})

		so.On("flixy play", func(sid string) {
			s, ok := sessions[sid]
			if !ok {
				log.Fatalf("`flixy play` from %s (%s) had an invalid session_id", so.Id(), so.Request().RemoteAddr)
			}

			s.Play()
		})

		// si -> session id
		so.On("flixy join", func(si string) {
			s, ok := sessions[si]
			if !ok {
				so.Emit("flixy invalid session id", si)
				return
			}

			s.AddMember(so)
		})

		so.On("flixy test", func(thing interface{}) {
			s.Sync()
		})

		log.Println("on connection")
		log.Printf("id %s connected", so.Id())
	})

	server.On("disconnection", func(so socketio.Socket) {
		sockid := so.Id()

		log.Printf("%v disconnected", sockid)
		log.Printf("deleting %v's memberships ....", sockid)
		for _, session := range sessions {
			session.RemoveMember(sockid)
		}
	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	mux := http.NewServeMux()
	api := routes.New()

	api.Get("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s\n", req.Header.Get("X-Request-ID"), req.URL.Path)
		fmt.Fprintf(w, "Hi!")
	})
	api.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		// print status here
		// this is just for debugging for now, we need more in-depth stuff soon
		enc := json.NewEncoder(w)
		enc.Encode(sessions)
	})
	api.Get("/sessions/:sid", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		session, present := sessions[params.Get(":sid")]
		if !present {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Location", session.GetNetflixURL())
		w.WriteHeader(302)
		routes.ServeJson(w, &session)
	})

	// TODO make some kind of simple RESTful (GET-only, though) API to
	// introspect the existing sessions (e.g. /sessions/125-112-521 ->
	// { vid, tid } for ease of redirecting to the correct netflix URI
	// ideally, going to /sessions/{sid} will 302 to the right place on
	// netflix (e.g. /watch/1243125?track_id=512312&flixy_id=125-112-521)
	// or something

	mux.Handle("/socket.io/", server)
	mux.Handle("/", api)

	n := negroni.Classic()
	middleware.Inject(n)
	n.UseHandler(mux)
	n.Run(":3000")
}
