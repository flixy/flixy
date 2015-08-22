package models

type GetSyncMessage struct {
	SessionID string `json:"session_id"`
}
type NewMessage struct {
	VideoID int    `json:"video_id"`
	Time    int    `json:"time"`
	Nick    string `json:"nick"`
}

type PauseMessage struct {
	SessionID string `json:"session_id"`
}

type PlayMessage struct {
	SessionID string `json:"session_id"`
}

type JoinMessage struct {
	SessionID string `json:"session_id"`
	Nick      string `json:"nick"`
}
