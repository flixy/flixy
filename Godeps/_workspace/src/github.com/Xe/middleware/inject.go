package middleware

import (
	"github.com/flixy/flixy/Godeps/_workspace/src/github.com/Xe/middleware/xff"
	"github.com/flixy/flixy/Godeps/_workspace/src/github.com/Xe/middleware/xrequestid"
	"github.com/flixy/flixy/Godeps/_workspace/src/github.com/codegangsta/negroni"
)

// Inject adds x-request-id and x-forwarded-for support to an existing negroni instance.
func Inject(n *negroni.Negroni) {
	n.Use(negroni.HandlerFunc(xff.XFF))
	n.Use(xrequestid.New(26))
}
