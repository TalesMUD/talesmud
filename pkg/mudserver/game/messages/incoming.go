package messages

// IncomingMessage is a client payload on the game socket.
// Classic play sends {"message":"..."}. Door mode also sends {"type":"door_key","key":"F"}.
type IncomingMessage struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Key     string `json:"key,omitempty"`
}
