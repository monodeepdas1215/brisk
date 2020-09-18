package brisk

import (
	"github.com/gobwas/ws"
	"net"
	"sync"
)

type clientsHub struct {
	sync.RWMutex
	clients				map[string]*socketClient

	// a common channel every socketClient listens to for a common broadcast functionality
	broadcastChannel	chan interface{}
}

// creates a new instance of hub with a max buffer size of broadcast channel passed as parameters
func newClientsHub(broadcastMessagesLimit int64) *clientsHub {
	return &clientsHub{
		clients:          make(map[string]*socketClient),
		broadcastChannel: make(chan interface{}, broadcastMessagesLimit),
	}
}


// get a particular client
func (ch *clientsHub) getClient(id string) *socketClient {

	var res *socketClient = nil

	ch.RLock()
	if val, ok := ch.clients[id]; ok {
		res = val
	}
	ch.RUnlock()

	return res
}

// adds a client to hub
func (ch *clientsHub) addClient(id string, socket *net.Conn) {

	cl := newSocketClient(id, socket)

	ch.Lock()
	ch.clients[id] = cl
	ch.Unlock()
}

// removes the client from hub and calls the callback function
func (ch *clientsHub) removeClient(id string, callbackFunc func(clientId string, err error)) {

	var client *socketClient = nil
	var err error = nil

	ch.Lock()
	if val, ok := ch.clients[id]; ok {
		client = val
		if err = val.closeConnection(); err != nil {
			AppLogger.logger.Errorf("error occurred while trying to close socket connection for client: %s  :  ", val.Id, err)
			AppLogger.logger.Errorf("could not remove client: %s", val.Id)
		} else {
			delete(ch.clients, val.Id)
		}
	}
	ch.Unlock()

	if callbackFunc != nil && client != nil {
		callbackFunc(client.Id, err)
	}
}

// finds the given client and pushes data to it
func (ch *clientsHub) pushToClient(clientId string, data []byte, opCode ws.OpCode) error {

	var client *socketClient = nil

	ch.RLock()

	if val, ok := ch.clients[clientId]; ok {
		client = val
	}
	ch.RUnlock()

	// TODO make this function call via threadpool
	if client != nil {
		go client.PushData(data, opCode)
	}
	return nil
}

// sets the authentication status of client to true
func (ch *clientsHub) authenticateClient(clientId string) {

	ch.RLock()
	if _, ok := ch.clients[clientId]; ok {
		ch.clients[clientId].SetAuthenticationStatus(true)
	}
	ch.RUnlock()
}