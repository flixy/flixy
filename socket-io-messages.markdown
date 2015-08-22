# socket.io messages present in flixy

## Messages clients can send

### `flixy pause`
#### Argument: `sid` (string)

Pauses the time in the given session ID.

#### Response:
	None specifically, but a `flixy sync` will be sent.

### `flixy play`
#### Argument: `sid` (string)

Plays the given session ID.

#### Response:
	None specifically, but a `flixy sync` will be sent.

### `flixy new`
#### Argument: ```{
	"video_id": int, // the current video ID
	"track_id": int, // the current track ID (????)
	"time": int, // the current timestamp according to netflix
	"nick": string // the client's preferred nickname
}``` (object)

Initializes a new session.

#### Response:
	A `flixy new session` response.

### `flixy join`
#### Argument: `sid` (string)

Joins the member to the given session.

#### Response:
	None specifically, however the user will be immediately synced with a `flixy sync` upon join.

## Messages the server can send

### `flixy invalid new init map`
#### Payload: the init map you sent with `flixy new`

### `flixy invalid session id`
#### Payload: the invalid session ID you sent with `flixy pause` or `flixy
play` or `flixy join`

### `flixy new session`
#### Payload: ```
{
	"session_id": string,
	"video_id": int,
	"track_id": int,
	"time": int,
	"paused": bool,
	"members": map[string]{
		"nick": string
	}
}
```

### `flixy join session`
#### Payload: ```
{
	"session_id": string,
	"video_id": int,
	"track_id": int,
	"time": int,
	"paused": bool,
	"members": map[string]{
		"nick": string
	}
}
```

### `flixy sync`
#### Payload: ```
{
	"video_id": int,
	"track_id": int,
	"time": int,
	"paused": bool
}
```
