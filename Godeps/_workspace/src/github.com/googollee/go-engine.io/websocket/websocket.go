package websocket

import (
	"github.com/flixy/flixy/Godeps/_workspace/src/github.com/googollee/go-engine.io/transport"
)

var Creater = transport.Creater{
	Name:      "websocket",
	Upgrading: true,
	Server:    NewServer,
	Client:    NewClient,
}
