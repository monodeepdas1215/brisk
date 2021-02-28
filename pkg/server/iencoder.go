package server

type IEncoder interface {
	Decode([]byte) (*Message, error)
	Encode(Message) ([]byte)
}


// Application Message format
type Message struct {
	Event	string					`json:"event"`
	Payload	map[string]interface{}	`json:"payload"`
}