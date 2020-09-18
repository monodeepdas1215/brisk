package brisk

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
)

// structure holding information about an incoming websocket connection
type socketClient struct {

	Id            string
	socket        *net.Conn
	authenticated bool
}

func newSocketClient(socketObj *net.Conn) *socketClient {
	return &socketClient{
		Id:            "",
		socket:        socketObj,
		authenticated: false,
	}
}

func (cl *socketClient) GetAuthenticationStatus() bool {
	return cl.authenticated
}

func (cl *socketClient) Authenticate(clientId string) {
	cl.Id = clientId
	cl.authenticated = true
}

func (cl *socketClient) RevokeAuthentication() {
	cl.authenticated = false
}

// pushes the data passed to the socketClient and returns error if any
func (cl *socketClient) PushData(data []byte, opCode ws.OpCode) error {
	return wsutil.WriteServerMessage(*cl.socket, opCode, data)
}

// close the socket connection and propagate any error
func (cl *socketClient) closeConnection() error {

	err := (*cl.socket).Close()
	if err != nil {
		return err
	}

	cl.socket = nil
	return nil
}