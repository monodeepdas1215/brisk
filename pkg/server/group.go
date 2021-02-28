package server

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/gobwas/ws"
	"sync"
)


type group struct {
	sync.RWMutex

	// group unique identifier
	Id	string

	// can hold any group metadata
	metadata map[string]interface{}

	// holds a connection to a client which is just connected and has not yet authenticated
	clients 			map[string]*socketClient

	pubsubChannel		*redis.PubSub

	// a common channel every socketClient listens to for a common broadcast functionality
	broadcastChannel	chan interface{}

	// channel to safely close all the connected clients and the group channels properly
	shutdownChannel		chan interface{}

	// downChannel forwards the data to down listeners for any intended purpose
	downChannel		chan interface{}
}

// creates a new instance of hub with a max buffer size of broadcast channel passed as parameters
func newGroup(groupId string, broadcastChannelCap int64) *group {

	g := &group{
		Id: groupId,
		clients:    make(map[string]*socketClient),
		broadcastChannel: make(chan interface{}, broadcastChannelCap),
		shutdownChannel: make(chan interface{}),
		downChannel: make(chan interface{}, broadcastChannelCap),
	}

	AppLogger.Infof("[newGroup] group: %s created", groupId)
	return g
}

func (g *group) getClient(id string) *socketClient {

	var res *socketClient = nil

	g.RLock()
	if val, ok := g.clients[id]; ok {
		res = val
	} else if val, ok := g.clients[id]; ok {
		res = val
	}
	g.RUnlock()

	return res
}

// adds a client to the common clients map since it is new and not yet authenticated
func (g *group) addClient(client *socketClient) {
	g.Lock()
	g.clients[client.Id] = client
	g.Unlock()
}

// finds the client from the clientHub, closes the socket connection and then removes it from clientHub
func (g *group) removeClient(id string, callbackFunc func(clientId string, err error)) {

	var client *socketClient = nil
	var err error = nil

	g.Lock()
	if val, ok := g.clients[id]; ok {
		client = val
		if err = val.StopClient(); err != nil {
			AppLogger.Errorf("[removeClient] client: %v: %v", val.Id, err)
		} else {
			AppLogger.Infof("[removeClient] client: %s closed\n", id)
		}
		delete(g.clients, val.Id)
	}
	g.Unlock()

	if callbackFunc != nil && client != nil {
		callbackFunc(client.Id, err)
	}
}

// finds the given client and pushes data to it
func (g *group) pushToClient(clientId string, data []byte, opCode ws.OpCode) error {

	var client *socketClient = nil

	g.RLock()
	if val, ok := g.clients[clientId]; ok {
		client = val
	}
	g.RUnlock()

	if client != nil {
		if err := client.PushData(data, opCode); err != nil {
			AppLogger.Errorln("[pushToClient] error occurred while pushing data to client: ", err)
		}
	}
	return nil
}

func (g *group) broadcastReceiver() {

	go func() {
		for {
			select {
			case incomingBroadcast:= <- g.broadcastChannel:

				dataStr := incomingBroadcast.(string)

				g.RLock()
				for key, _ := range g.clients {
					g.clients[key].PushData([]byte(dataStr), ws.OpText)
				}
				g.RUnlock()

			case <-g.shutdownChannel:
				AppLogger.Infoln("[broadcastReceiver] shutting down")
				return
			}
		}
	}()
}

func (g *group) createPubSubConnection(channelName string) {

	if g.pubsubChannel == nil {
		g.pubsubChannel = RedisClient.Subscribe(context.Background(), channelName)

		if err := g.pubsubChannel.Subscribe(context.Background(), channelName); err != nil {
			AppLogger.Errorf("[createPubSubConnection] err: %v", err)
		}
	}
}


func (g *group) createBroadcast(msg interface{}) {
	g.broadcastChannel <- msg
}