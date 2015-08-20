// Package main provides the executable flixy server
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/Xe/middleware"
	"github.com/codegangsta/negroni"
	"github.com/drone/routes"
	"github.com/googollee/go-socket.io"
)

// sessions is a map of some identifier to each Flixy session.
// TODO figure out what to make this identifier
var sessions map[string]*Session = make(map[string]*Session)
var socketlist map[string]*socketio.Socket = make(map[string]*socketio.Socket)

func makeNewSessionId() string {
	return fmt.Sprintf("%4d-%4d-%4d-%4d", rand.Intn(9999), rand.Intn(9999), rand.Intn(9999), rand.Intn(9999))
}

func main() {
	fmt.Println("Hello, world!")

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	sessions["test"] = &Session{SessionID: "test", VideoID: 1, TrackID: 2, Time: 3, Members: make(map[string]*Member)}

	// TODO figure out what the fuck is the deal with IDs --- can they be a
	// key in map[session_id]User or something?
	server.On("connection", func(so socketio.Socket) {
		socketlist[so.Id()] = &so
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
			sid := makeNewSessionId()

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

			s := &Session{
				SessionID: sid,
				VideoID:   vid,
				TrackID:   tid,
				Time:      time,
				Members:   make(map[string]*Member),
			}

			sessions[sid] = s

			so.Emit("flixy new session", *s)
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
		log.Printf("id %s connected, currently connected:", so.Id())
		for k, v := range socketlist {
			log.Printf("%v -> %v", k, v)
		}
	})

	server.On("disconnection", func(so socketio.Socket) {
		delete(socketlist, so.Id())
		sockid := so.Id()

		log.Printf("%v disconnected, connected now:", sockid)
		for k, v := range socketlist {
			log.Printf("%v -> %v", k, v)
		}
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
