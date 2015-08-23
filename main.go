// Package main provides the executable flixy server
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"math/rand"
	"net/http"

	log "github.com/Sirupsen/logrus"
	flag "github.com/ogier/pflag"

	"github.com/flixy/flixy/models"

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
	opts     = options{}
	loglevel log.Level
)

// sessions is a map of the session identifier to each Flixy session, generated
// by `makeNewSessionID`
var sessions = make(map[string]*models.Session)

// members is a map of socket identifier to member, for ease of removing users
// from sessions after they disconnect and other such happenstances
var members = make(map[string]*models.Member)

// makeNewSessionID produces a session identifier, which is currently of the
// form "%4d-%4d-%4d-%4d" but this is subject to change and is an
// implementation detail.
func makeNewSessionID() string {
	return fmt.Sprintf("%4d-%4d-%4d-%4d", rand.Intn(9999), rand.Intn(9999), rand.Intn(9999), rand.Intn(9999))
}

// init is a special function called before main(). used to set up such things
// as loglevels and such.
func init() {
	defaultPort, err := strconv.Atoi(os.Getenv("FLIXY_PORT"))
	if err != nil {
		defaultPort = 3000
	}
	defaultHost := os.Getenv("FLIXY_HOST")
	if defaultHost == "" {
		defaultHost = "0.0.0.0"
	}

	defaultLogLevel := os.Getenv("FLIXY_LOGLEVEL")
	if defaultLogLevel == "" {
		defaultLogLevel = "info"
	}

	flag.IntVarP(&opts.Port, "port", "p", defaultPort, "the port to listen on")
	flag.StringVarP(&opts.Host, "host", "H", defaultHost, "the host to listen on")
	flag.StringVarP(&opts.LogLevel, "log-level", "l", defaultLogLevel, "the log level to use (possible: panic,fatal,error,warn,info,debug)")
	flag.Parse()

	ll, ok := logLevels[opts.LogLevel]
	if !ok {
		log.Errorf("invalid log level %s set, falling back to default %s", opts.LogLevel, "info")
		ll = log.InfoLevel
	}

	loglevel = ll
	log.SetLevel(loglevel)
	log.Debugf("setting log level to %s", opts.LogLevel)

}

// main is the entry point to the flixy server.
func main() {
	log.Info("Starting flixy!")

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// TODO figure out what the fuck is the deal with IDs --- can they be a
	// key in map[session_id]User or something?
	server.On("connection", func(so socketio.Socket) {
		sockid := so.Id()
		sockip := so.Request().RemoteAddr

		so.On("flixy get sync", SyncHandler(so))
		so.On("flixy new", NewHandler(so))
		so.On("flixy pause", PauseHandler(so))
		so.On("flixy play", PlayHandler(so))
		so.On("flixy join", JoinHandler(so))

		log.WithFields(log.Fields{
			"member_sockid": sockid,
			"member_remote": sockip,
		}).Info("connected")
	})

	server.On("disconnection", func(so socketio.Socket) {
		sockid := so.Id()
		sockip := so.Request().RemoteAddr

		m, ok := members[sockid]
		if !ok {
			log.WithFields(log.Fields{
				"verb":          "disconnection",
				"member_sockid": sockid,
				"member_remote": sockip,
			}).Warn("member never in a session disconnected")
		}

		log.Infof("%v disconnected", sockid)
		delete(members, sockid)

		s := m.Session
		s.RemoveMember(sockid)
	})

	server.On("error", func(so socketio.Socket, err error) {
		// TODO how can this even happen?
		log.Error("error:", err)
	})

	mux := http.NewServeMux()
	api := routes.New()

	// TODO remove this
	api.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		// print status here
		// this is just for debugging for now, we need more in-depth stuff soon
		enc := json.NewEncoder(w)
		wms := make(map[string]models.WireSession)

		for k, v := range sessions {
			wms[k] = v.GetWireSession()
		}
		enc.Encode(wms)
	})

	// `/sessions/:sid` will 302 the user to the proper Netflix URL if it's
	// a valid SID, setting the session ID in the URL as it does so. It
	// will *also* return the session as a JSON object.
	api.Get("/sessions/:sid", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		session, present := sessions[params.Get(":sid")]
		if !present {
			w.WriteHeader(404)
			return
		}

		w.Header().Set("Location", session.GetNetflixURL())
		w.WriteHeader(302)
		routes.ServeJson(w, session.GetWireSession())
	})

	mux.Handle("/socket.io/", server)
	mux.Handle("/", api)

	n := negroni.New()
	n.Use(negronilogrus.NewCustomMiddleware(loglevel, &log.TextFormatter{}, "web"))
	n.Use(negroni.NewRecovery())
	middleware.Inject(n)
	n.UseHandler(mux)
	n.Run(fmt.Sprintf("%s:%d", opts.Host, opts.Port))
}
