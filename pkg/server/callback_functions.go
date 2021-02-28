package server

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

func NewServerCallbacks(onClientConnected func(clientId string),
	authHandler func(clientId string, msg Message) (bool, string),
	onMessageReceived func(clientId string, msg Message, err error),
	onClientDisconnected func(clientId string, err error),
	onHostConnectHandler func(host []byte) error,
	onHeaderHandler func(key, value []byte) (err error),
	onBeforeUpgrade func() (header ws.HandshakeHeader, err error)) *ServerCallbacks {

	callback := &ServerCallbacks{}

	if onClientConnected == nil {
		callback.OnClientConnected = DefaultOnClientConnected
	} else {
		callback.OnClientConnected = onClientConnected
	}

	if authHandler == nil {
		callback.AuthHandler = DefaultAuthHandler
	} else {
		callback.AuthHandler = authHandler
	}

	if onMessageReceived == nil {
		callback.OnMessageReceived = DefaultOnMessageReceived
	} else {
		callback.OnMessageReceived = onMessageReceived
	}

	if onClientDisconnected == nil {
		callback.OnClientDisconnected = DefaultOnClientDisconnected
	} else {
		callback.OnClientDisconnected = onClientDisconnected
	}

	if onHostConnectHandler == nil {
		callback.OnHostConnectHandler = DefaultOnHostConnectHandler
	} else {
		callback.OnHostConnectHandler = onHostConnectHandler
	}

	if onHeaderHandler == nil {
		callback.OnHeaderHandler = DefaultOnHeaderHandler
	} else {
		callback.OnHeaderHandler = onHeaderHandler
	}

	if onBeforeUpgrade == nil {
		callback.OnBeforeUpgrade = DefaultOnBeforeUpgrade
	} else {
		callback.OnBeforeUpgrade = onBeforeUpgrade
	}

	return callback
}

func DefaultOnClientConnected(clientId string) {
	AppLogger.Infof("[DefaultOnClientConnected] new client connected: %s", clientId)
}

func DefaultAuthHandler(clientId string, msg Message) (bool, string) {
	AppLogger.Infof("[DefaultAuthHandler] default auth handler being used, doing nothing except forwarding requests")
	return true, "default auth handler"
}

func DefaultOnMessageReceived(clientId string, msg Message, err error) {
	if err != nil {
		AppLogger.Errorf("[DefaultOnMessageReceived] error occurred while getting message from client: %s says %s --- %v",
			clientId, msg, err)
	}
	AppLogger.Infof("[DefaultOnMessageReceived] client: %s says %s", clientId, msg)
}

func DefaultOnClientDisconnected(clientId string, err error) {
	if err != nil {
		AppLogger.Errorf("[DefaultOnClientDisconnected] err occurred while client disconnection: %v", err)
	} else {
		AppLogger.Infof("[DefaultOnClientDisconnected] client: %s disconnected", clientId)
	}
}

func DefaultOnHostConnectHandler(host []byte) error {
	AppLogger.Infof("[DefaultOnHostConnectHandler] host: %s", string(host))
	return nil
}

func DefaultOnHeaderHandler(key, value []byte) error {
	AppLogger.Infof("[DefaultOnHeaderHandler] Header[%s] = %s", string(key), string(value))
	return nil
}

func DefaultOnBeforeUpgrade() (header ws.HandshakeHeader, err error) {
	AppLogger.Infoln("[DefaultOnBeforeUpgrade] using default onBeforeUpgrade handler")
	return header, err
}
