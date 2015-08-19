// Package main provides the executable flixy server
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Xe/middleware"
	"github.com/codegangsta/negroni"
	"github.com/googollee/go-socket.io"
)

// sessions is a map of some identifier to each Flixy session.
// TODO figure out what to make this identifier
var sessions map[string]Session = make(map[string]Session)
var socketlist map[string]socketio.Socket = make(map[string]socketio.Socket)

func main() {
	fmt.Println("Hello, world!")

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	sessions["test"] = Session{VideoID: 1, TrackID: 2, Time: 3, Members: make(map[string]*Member)}

	// TODO figure out what the fuck is the deal with IDs --- can they be a
	// key in map[session_id]User or something?
	server.On("connection", func(so socketio.Socket) {
		socketlist[so.Id()] = so
		s := sessions["test"]
		s.Members[so.Id()] = &Member{so}
		log.Printf("34 %v", s)

		// se -> sync event
		// map[string]interface{} that looks like
		// { paused: false, ts: 128519241124, vid: 6527312, tid: 5124623 }
		so.On("flixy sync", func(se map[string]interface{}) {
			log.Printf("%v", se)
		})

		so.On("flixy new", func(nse map[string]interface{}) {
			// does this need to be a separate handler? surely we can merge this and `join`
		})

		// si -> session id
		so.On("flixy join", func(si string) {
			// hmm! we could probably merge this and `flixy new`.
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

		log.Printf("%v disconnected, connected now:", so.Id())
		for k, v := range socketlist {
			log.Printf("%v -> %v", k, v)
		}
		log.Printf("deleting %v's memberships ....")
		for _, session := range sessions {
			session.RemoveMember(so.Id())
		}
	})

	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s\n", req.Header.Get("X-Request-ID"), req.URL.Path)
		fmt.Fprintf(w, "Hi!")
	})
	mux.Handle("/socket.io/", server)

	n := negroni.Classic()
	middleware.Inject(n)
	n.UseHandler(mux)
	n.Run(":3000")
}
