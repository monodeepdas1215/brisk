package core

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
)


// handles the entire cycle of communication for a websocket connection after it is connected to the server
type handleMessagesWorkReq struct {
	client *socketClient
	ss	 *SocketServer
}

func (hm *handleMessagesWorkReq) Execute() {

	defer func() {
		hm.ss.pushToClient(hm.client.Id, nil, ws.OpClose)
		hm.ss.removeClient(hm.client.Id, hm.ss.OnClientDisconnected)
	}()

	// waiting for the messages from client
	for {

		// reading message in bytes
		message, op, err := wsutil.ReadClientData(*hm.client.socket)
		if err != nil {
			AppLogger.logger.Errorln("error occurred while reading client data: ", err)
			break
		}

		if op == ws.OpClose {
			hm.ss.SendMessage(false, hm.client.Id, Message{}, ws.OpClose)
			break
		}

		if op == ws.OpPing {
			hm.ss.SendMessage(false, hm.client.Id, Message{}, ws.OpPong)
			continue
		}

		// converting bytes into configured message type
		decodedMessage, err := hm.ss.encoder.Decode(message)
		if err != nil {

			// notify the client about the error
			hm.ss.AddWorkRequest(&serverReplyWorkReq{
				obj: hm.client.socket,
				msg: &Message{
					Event:   "error",
					Payload: map[string]interface{}{
						"code": http.StatusBadRequest,
						"message": err.Error(),
					},
				},
				ss:  hm.ss,
			})
			continue
		}

		// if auth is configured and client not authenticated then perform authentication
		if hm.ss.AuthHandler != nil && !hm.client.GetAuthenticationStatus() {

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

				hm.ss.sendAcknowledgement(hm.client.socket, msg)
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

			if hm.ss.SendAcknowledgement {
				// callback
				hm.ss.AddWorkRequest(&serverReplyWorkReq{
					obj: hm.client.socket,
					msg: &Message{
						Event:   "ack",
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
	id  string
	msg Message
	err error
	ss  *SocketServer
}

func (mr *msgReceivedCallbackWorkReq) Execute() {
	mr.ss.OnMessageReceived(mr.id, mr.msg, mr.err)
}

func (mr *msgReceivedCallbackWorkReq) GetId() string {
	return "message received callback"
}

// work request to send message to client
type sendMessageWorkReq struct {
	clientId string
	code     ws.OpCode
	ss       *SocketServer
	msg      Message
}

func (sm *sendMessageWorkReq) Execute() {
	data := sm.ss.encoder.Encode(sm.msg)
	sm.ss.pushToClient(sm.clientId, data, sm.code)
}

func (sm *sendMessageWorkReq) GetId() string {
	return "send_message_work_req"
}