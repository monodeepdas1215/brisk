package brisk

import "github.com/gobwas/ws"

type ServerCallbacks struct {

	// called after client is connected to the server
	OnClientConnected		func(clientId string)

	// authentication handler callback
	AuthHandler				func(clientId string, msg Message) (bool, string)

	// on message is received
	OnMessageReceived		func(clientId string, msg Message, err error)

	// after the client is disconnected
	OnClientDisconnected	func(clientId string, err error)

	// this function can be used to apply some custom logic when the client is connected to the server
	OnHostConnectHandler	func(host []byte) error

	// function to apply custom logic to handle headers upon client connection
	OnHeaderHandler			func(key, value []byte) (err error)

	// called just before connection is upgraded to websocket
	OnBeforeUpgrade			func() (header ws.HandshakeHeader, err error)
}
