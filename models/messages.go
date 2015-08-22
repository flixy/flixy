package models

// these are the internal message structs that JSON gets unmarshaled into.
// please do not depend on them, aside from client authors structuring their
// JSON data around it

// GetSyncMessage is the struct to which `flixy get sync` messages are
// unmarshaled into.
type GetSyncMessage struct {
	SessionID string `json:"session_id"`
}

// NewMessage is the struct to which `flixy new` messages are unmarshaled into.
type NewMessage struct {
	VideoID int    `json:"video_id"`
	Time    int    `json:"time"`
	Nick    string `json:"nick"`
}

// PauseMessage is the struct to which `flixy pause` messages are unmarshaled
// into.
type PauseMessage struct {
	SessionID string `json:"session_id"`
}

// PlayMessage is the struct to which `flixy play` messages are unmarshaled
// into.
type PlayMessage struct {
	SessionID string `json:"session_id"`
}

// JoinMessage is the struct to which `flixy join` messages are unmarshaled
// into.
type JoinMessage struct {
	SessionID string `json:"session_id"`
	Nick      string `json:"nick"`
}
