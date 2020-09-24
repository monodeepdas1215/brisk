package core

import (
	"github.com/gobwas/ws"
	"sync"
)

type clientsHub struct {
	sync.RWMutex
	// holds a connection to a client which is just connected and has not yet authenticated
	commonClients 			map[string]*socketClient

	// holds a client after it is has successfully authenticated
	authenticatedClients 	map[string]*socketClient

	// a common channel every socketClient listens to for a common broadcast functionality
	broadcastChannel		chan interface{}
}

// creates a new instance of hub with a max buffer size of broadcast channel passed as parameters
func newClientsHub(broadcastMessagesLimit int64) *clientsHub {
	return &clientsHub{
		commonClients:    make(map[string]*socketClient),
		authenticatedClients: make(map[string]*socketClient),
		broadcastChannel: make(chan interface{}, broadcastMessagesLimit),
	}
}


// get a particular client
func (ch *clientsHub) getClient(id string) *socketClient {

	var res *socketClient = nil

	ch.RLock()
	if val, ok := ch.commonClients[id]; ok {
		res = val
	} else if val, ok := ch.authenticatedClients[id]; ok {
		res = val
	}
	ch.RUnlock()

	return res
}

// adds a client to the common clients map since it is new and not yet authenticated
func (ch *clientsHub) addClient(client *socketClient) {

	ch.Lock()
	ch.commonClients[client.Id] = client
	ch.Unlock()
}

// finds the client from the clientHub, closes the socket connection and then removes it from clientHub
func (ch *clientsHub) removeClient(id string, callbackFunc func(clientId string, err error)) {

	var client *socketClient = nil
	var err error = nil

	ch.Lock()
	if val, ok := ch.commonClients[id]; ok {
		client = val
		if err = val.closeConnection(); err != nil {
			AppLogger.logger.Errorf("error occurred while trying to close socket connection for client: %s  :  ", val.Id, err)
			AppLogger.logger.Errorf("could not remove client: %s", val.Id)
		}
		delete(ch.commonClients, val.Id)
	} else if val, ok := ch.authenticatedClients[id]; ok {
		client = val
		if err = val.closeConnection(); err != nil {
			AppLogger.logger.Errorf("error occurred while trying to close socket connection for client: %s  :  ", val.Id, err)
			AppLogger.logger.Errorf("could not remove client: %s", val.Id)
		}
		delete(ch.authenticatedClients, val.Id)
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

	if val, ok := ch.commonClients[clientId]; ok {
		client = val
	} else if val, ok := ch.authenticatedClients[clientId]; ok {
		client = val
	}
	ch.RUnlock()

	if client != nil {
		if err := client.PushData(data, opCode); err != nil {
			AppLogger.logger.Errorln("error occurred while pushing data to client: ", err)
		}
	}
	return nil
}

// removes client from the un authenticated map to authenticated map and assigns it with the given client id
func (ch *clientsHub) authenticateClient(oldClientId, newClientId string) {

	ch.RLock()
	if val, ok := ch.commonClients[oldClientId]; ok {

		// move the client to authenticated client
		ch.authenticatedClients[newClientId] = val
		val.Authenticate(newClientId)

		// delete from unAuthenticated Clients
		delete(ch.commonClients, oldClientId)
	}
	ch.RUnlock()
}