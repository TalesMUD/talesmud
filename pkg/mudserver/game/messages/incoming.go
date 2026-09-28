package messages

// IncomingMessage is a client payload on the game socket.
// Classic play sends {"message":"..."}. The text client also sends {"type":"door_key","key":"n"}.
type IncomingMessage struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Key     string `json:"key,omitempty"`
}
