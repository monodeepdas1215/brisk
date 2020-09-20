package brisk

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
)

// work request to callback the on client connected function
type onClientConnectedWorkReq struct {
	id string
	ss *SocketServer
}

func (occ *onClientConnectedWorkReq) Execute() {
	if occ.ss.OnClientConnected != nil {
		occ.ss.OnClientConnected(occ.id)
	}
}

func (occ *onClientConnectedWorkReq) GetId() string {
	return "on client connected work req"
}


// handles the entire cycle of communication for a websocket connection after it is connected to the server
type handleMessagesWorkReq struct {
	client *socketClient
	ss	 *SocketServer
}

func (hm *handleMessagesWorkReq) Execute() {

	defer hm.ss.removeClient(hm.client.Id, hm.ss.OnClientDisconnected)

	// waiting for the messages from client
	for {

		// reading message in bytes
		message, op, err := wsutil.ReadClientData(*hm.client.socket)
		if err != nil {
			AppLogger.logger.Errorln("error occurred while reading client data: ", err)
			break
		}

		if op == ws.OpClose {
			break
		}

		if op == ws.OpPing {
			// TODO implement a ping pong functionality
		}

		AppLogger.logger.Debugln("message from client: ", string(message))

		// converting bytes into configured message type
		decodedMessage, err := hm.ss.encoder.Decode(message)
		if err != nil {

			hm.ss.AddWorkRequest(&serverReplyWorkReq{
				obj: hm.client.socket,
				msg: &Message{
					Event:   "error",
					Payload: map[string]interface{}{
						"code": http.StatusInternalServerError,
						"message": err.Error(),
					},
				},
				ss:  hm.ss,
			})
			break
		}

		// if auth is configured and client not authenticated then perform authentication
		if hm.ss.AuthenticationType != AUTH_TYPE_NO_AUTH && !hm.client.GetAuthenticationStatus() {

			authResult, clientId := hm.ss.AuthHandler(hm.client.Id, *decodedMessage)
			if authResult {
				hm.ss.authenticateClient(hm.client.Id, clientId)

			} else {

				// break the connection if authentication is failed
				msg := &Message{
					Event:   "error",
					Payload: map[string]interface{}{
						"code": http.StatusUnauthorized,
						"message": "could not authenticate client",
					},
				}
				hm.ss.AddWorkRequest(&serverReplyWorkReq{
					obj: hm.client.socket,
					msg: msg,
					ss:  hm.ss,
				})
				break
			}

		} else {

			// if either authentication is not configured or the client is authenticated then receive the message
			hm.ss.AddWorkRequest(&msgReceivedCallbackWorkReq{
				id:  hm.client.Id,
				msg: *decodedMessage,
				err: err,
				ss:  hm.ss,
			})

			// callback
			hm.ss.AddWorkRequest(&serverReplyWorkReq{
				obj: hm.client.socket,
				msg: &Message{
					Event:   "reply",
					Payload: map[string]interface{}{
						"code": 200,
						"message": "ok",
					},
				},
				ss:  hm.ss,
			})
		}
	}
}

func (hm *handleMessagesWorkReq) GetId() string {
	return hm.client.Id + " interaction cycle"
}


// work request to send server replies
type serverReplyWorkReq struct {
	obj *net.Conn
	msg *Message
	ss *SocketServer
}

func (ue *serverReplyWorkReq) Execute() {
	ue.ss.sendAcknowledgement(ue.obj, ue.msg)
}

func (ue *serverReplyWorkReq) GetId() string {
	return "server reply"
}


// work request to call back the function
type msgReceivedCallbackWorkReq struct {
	id string
	msg Message
	err error
	ss *SocketServer
}

func (mr *msgReceivedCallbackWorkReq) Execute() {
	mr.ss.OnMessageReceived(mr.id, mr.msg, mr.err)
}

func (mr *msgReceivedCallbackWorkReq) GetId() string {
	return "message received callback"
}