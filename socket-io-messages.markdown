# socket.io messages present in flixy

## Messages clients can send

All messages that clients send *MUST* be encoded with JSON.stringify before
being sent.

### `flixy get sync`
#### Argument: `{ "session_id": string }`

Asks the server to send a sync update.

### `flixy pause`
#### Argument: `{ "session_id": string }`

Pauses the time in the given session ID.

#### Response:
	None specifically, but a `flixy sync` will be sent.

### `flixy play`
#### Argument: `{ "session_id": string }`

Plays the given session ID.

### `flixy seek`
#### Argument: `{ "session_id": string, "time": int }`

Sets the given session to the given timestamp

#### Response:
	None specifically, but a `flixy sync` will be sent.

### `flixy new`
#### Argument: ` { "video_id": int, "time": int }`

Initializes a new session.

#### Response:
	A `flixy new session` response.

### `flixy join`
#### Argument: `{ "session_id": string, "nick": string }`

Joins the member to the given session.

#### Response:
	None specifically, however the user will be immediately synced with a `flixy sync` upon join.

### `flixy leave`
#### Argument: `{}`

Makes the member leave their current session. No further messages should be
sent.

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
	"session_id": string,
	"video_id": int,
	"time": int,
	"paused": bool,
	"members": map[string]{
		"nick": string
	}
}
```
