package server

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
)

// structure holding information about an incoming websocket connection
type socketClient struct {

	Id                   string
	metadata			 map[string]interface{}
	socket               *net.Conn
	broadCastReceiveChan chan interface{}
	stopBroadcastChan	 chan interface{}
}

func newSocketClient(tmpId string, socketObj *net.Conn) *socketClient {
	client := &socketClient{
		Id:            tmpId,
		socket:        socketObj,
		broadCastReceiveChan: make(chan interface{}, 20),
		stopBroadcastChan: make(chan interface{}),
	}
	return client
}

func (cl *socketClient) SetMetadata(meta map[string]interface{}) {
	cl.metadata = meta
}

func (cl *socketClient) GetMetadata(key string) interface{} {

	val, ok := cl.metadata[key]
	if !ok {
		return nil
	}
	return val
}

// subscription goroutine for receiving group broadcasts
func (cl *socketClient) ListenToGroupBroadcast() {

	go func() {
		for {
			select {
			case groupBroadcastData := <- cl.broadCastReceiveChan:
				AppLogger.logger.Infof("[SubscribeForGroupBroadcast] broadcast received: %v", groupBroadcastData)

				groupBroadcastDataStr := groupBroadcastData.(string)
				cl.PushData([]byte(groupBroadcastDataStr), ws.OpText)

			case <-cl.stopBroadcastChan:
				AppLogger.logger.Infoln("[SubscribeForGroupBroadcast] exiting broadcast")
				return
			}
		}
	}()
}

// pushes the data passed to the socketClient and returns error if any
func (cl *socketClient) PushData(data []byte, opCode ws.OpCode) error {
	return wsutil.WriteServerMessage(*cl.socket, opCode, data)
}

// close all the allocated resource to the client
func (cl *socketClient) StopClient() error {

	cl.stopBroadcastChan <- 1

	close(cl.stopBroadcastChan)
	close(cl.broadCastReceiveChan)

	cl.PushData([]byte("closing connection from server"), ws.OpClose)

	err := (*cl.socket).Close()
	if err != nil {
		return err
	}
	return nil
}
