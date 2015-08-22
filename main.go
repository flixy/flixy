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

	log.Printf("Starting flixy!")

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// TODO figure out what the fuck is the deal with IDs --- can they be a
	// key in map[session_id]User or something?
	server.On("connection", func(so socketio.Socket) {
		so.On("flixy get sync", func(sid string) {
			s, ok := sessions[sid]
			if !ok {
				log.Fatalf("`flixy get sync` from %s (%s) had an invalid session_id", so.Id(), so.Request().RemoteAddr)
				so.Emit("flixy invalid session id", sid)
			}

			so.Emit("flixy sync", s.GetWireStatus())
		})

		so.On("flixy new", func(nse map[string]int) {
			sid := makeNewSessionID()

			vid, ok := nse["video_id"]
			if !ok {
				log.Fatalf("`flixy new` from %s (%s) had no video_id", so.Id(), so.Request().RemoteAddr)
				so.Emit("flixy invalid new init map", nse)
			}

			tid, ok := nse["track_id"]
			if !ok {
				log.Fatalf("`flixy new` from %s (%s) had no track_id", so.Id(), so.Request().RemoteAddr)
				so.Emit("flixy invalid new init map", nse)
			}

			time, ok := nse["time"]
			if !ok {
				log.Fatalf("`flixy new` from %s (%s) had no time", so.Id(), so.Request().RemoteAddr)
				so.Emit("flixy invalid new init map", nse)
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
				so.Emit("flixy invalid session id", sid)
			}

			s.Pause()
		})

		so.On("flixy play", func(sid string) {
			s, ok := sessions[sid]
			if !ok {
				log.Fatalf("`flixy play` from %s (%s) had an invalid session_id", so.Id(), so.Request().RemoteAddr)
				so.Emit("flixy invalid session id", sid)
			}

			s.Play()
		})

		// sid -> session id
		so.On("flixy join", func(sid string) {
			s, ok := sessions[sid]
			if !ok {
				so.Emit("flixy invalid session id", sid)
				return
			}

			s.AddMember(so)
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

	n := negroni.Classic()
	middleware.Inject(n)
	n.UseHandler(mux)
	n.Run(":3000")
}
