package main

import (
	"fmt"
	"net/http"

	"github.com/flixy/flixy/Godeps/_workspace/src/github.com/codegangsta/negroni"
	"github.com/flixy/flixy/Godeps/_workspace/src/github.com/meatballhat/negroni-logrus"
)

func main() {
	r := http.NewServeMux()
	r.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "success!\n")
	})

	n := negroni.New()
	n.Use(negronilogrus.NewMiddleware())
	n.UseHandler(r)

	n.Run(":9999")
}
