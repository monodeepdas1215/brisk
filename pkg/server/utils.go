package server

func AuthenticateMessage(msg *Message) bool {
	if msg.Event == AuthenticationEvent {
		return true
	}
	return false
}
