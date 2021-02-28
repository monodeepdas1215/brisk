package server

import (
	"github.com/go-redis/redis/v8"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"io"
	"net"
	"sync"
	"syscall"
)

type SocketServer struct {
	sync.RWMutex
	*Configuration
	*ServerCallbacks
	encoder IEncoder
	pubSubChan	*redis.PubSub
	groups  map[string]*group
}

// starts with a basic configuration setting
func DefaultSocketServer(config *Configuration) *SocketServer {

	InitLogger(config.LogLevel, config.LoggerReportCaller)

	obj := &SocketServer{
		Configuration:   config,
		groups: make(map[string]*group),
		//group:      	newGroup(config.BroadcastMessagesLimit),
		ServerCallbacks: NewServerCallbacks(nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil),
	}

	var appEncoder IEncoder

	switch obj.AcceptMessageEncoding {
	case ENCODING_TYPE_JSON:
		appEncoder = &JsonEncoder{}
	case ENCODING_TYPE_MSG_PACK:
		panic("encoding using msg_pack has not been implemented yet")
	}
	obj.encoder = appEncoder

	// seeding every server instance with a default group
	obj.AddGroup("default", config.BroadcastMessagesLimit)

	return obj
}

// Adds a group if not exists
func (ss *SocketServer) AddGroup(groupId string, groupBroadcastChannelLimit int64) {

	ss.Lock()
	if _, ok := ss.groups[groupId]; !ok {
		group := newGroup(groupId, groupBroadcastChannelLimit)
		group.createPubSubConnection(groupId)
		ss.groups[groupId] = group
	}
	ss.Unlock()
}

func (ss *SocketServer) AddClient(groupId string, clientObj *socketClient) {
	ss.Lock()
	if val, ok := ss.groups[groupId]; ok {
		val.addClient(clientObj)
	}
	ss.Unlock()
}

// does some initial checks to ensure the presence of all handler functions in place.
// increases the ulimit and starts the server loop
func (ss *SocketServer) StartListening() {

	// TODO increase ulimit
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(err)
	}

	// finally starting the server loop
	ss.startServerLoop()
}

func (ss *SocketServer) startServerLoop() {

	listener, err := ss.setupTCPConnection()
	if err != nil {
		return // stop the loop
	}

	// starting the blocking loop here
	for {

		// accepting a new connection
		conn, err := listener.Accept()
		if err != nil {
			AppLogger.Errorln("[startServerLoop] error occurred while listening to next connection: ", err)
			return
		}

		AppLogger.Infoln("incoming connection from: ", conn.RemoteAddr())

		// upgrade the tcp connection to websocket protocol
		_, err = ss.tcpConnectionUpgrader().Upgrade(conn)
		if err != nil {
			AppLogger.Errorln("[startServerLoop] error occurred while upgrading TCP connection to websocket: ", err)
			continue
		}

		ss.handleMessages(conn)

		//id := uuid.New().String()
		//incomingClient := newSocketClient(id, &conn)
		//
		//if ss.OnClientConnected != nil {
		//	ss.OnClientConnected(id)
		//}
	}
}

// setups the tcp connection for the given address
func (ss *SocketServer) setupTCPConnection() (net.Listener, error) {

	listener, err := net.Listen("tcp", ss.HostAddr)
	if err != nil {
		AppLogger.Errorln("[setupTCPConnection] error occurred while listening to TCP connections: ", err)
		return nil, err
	}

	AppLogger.Infoln("[setupTCPConnection] server listening for TCP connections on the address: ", ss.HostAddr)
	return listener, nil
}

// upgrades a tcp connection to websocket protocol
func (ss *SocketServer) tcpConnectionUpgrader() *ws.Upgrader {

	// initializing these functions if they are not initialized by the user
	if ss.OnHostConnectHandler == nil {
		ss.OnHostConnectHandler = func(host []byte) error {
			return nil
		}
	}
	if ss.OnHeaderHandler == nil {
		ss.OnHeaderHandler = func(key, value []byte) (err error) {
			return
		}
	}
	if ss.OnBeforeUpgrade == nil {
		ss.OnBeforeUpgrade = func() (header ws.HandshakeHeader, err error) {
			return header, nil
		}
	}

	return &ws.Upgrader{
		// Read and Write Buffer Sizes are set to default
		ReadBufferSize:  0,
		WriteBufferSize: 0,
		Header:          nil,
		OnRequest:       nil,
		OnHost: 		 ss.OnHostConnectHandler,
		OnHeader: 		 ss.OnHeaderHandler,
		OnBeforeUpgrade: ss.OnBeforeUpgrade,
	}
}

func (ss *SocketServer) handleMessages(conn net.Conn) {
	go func() {

		isClientAuthenticated := false

		for {

			header, err := ws.ReadHeader(conn)
			if err != nil {
				// handle error
				AppLogger.Errorf("[handleMessages] err occurred while reading header: %v", header)
			} else {
				AppLogger.Debugf("[handleMessages] headers: %+v", header)
			}

			if header.OpCode == ws.OpClose {
				if err := conn.Close(); err != nil {
					AppLogger.Errorf("[handleMessages] error occurred while closing connections: %v", err)
				}
				return
			}

			payload := make([]byte, header.Length)
			_, err = io.ReadFull(conn, payload)
			if err != nil {
				// handle error
				AppLogger.Errorf("[handleMessages] error occurred while reading payload: %v", err)
			}
			if header.Masked {
				ws.Cipher(payload, header.Mask, 0)
			}

			// process requests here
			message, _ := ss.encoder.Decode(payload)

			if !isClientAuthenticated {

				id, _ := uuid.NewUUID()
				clientObj := newSocketClient(id.String(), &conn)

				isClientAuthenticated, _ = ss.AuthHandler(id.String(), *message)

				// subscribing a client to a specific group
				groupId := message.Payload["group"].(string)
				if groupId != "" {
					// attach to default group
					ss.AddGroup(groupId, 100)
					ss.AddClient(groupId, clientObj)
				} else {
					ss.AddClient("default", clientObj)
				}

			} else {
			}

			// Reset the Masked flag, server frames must not be masked as
			// RFC6455 says.
			header.Masked = false

			if err := ws.WriteHeader(conn, header); err != nil {
				// handle error
				AppLogger.Errorf("[handleMessages] error occurred while writing header: %v", err)
			}
			if _, err := conn.Write(payload); err != nil {
				// handle error
				AppLogger.Errorf("[handleMessages] error occurred while writing payload: %v", err)
			}
		}
	}()
}

// if enabled this function sends an acknowledgment back to the connected client upon receiving data
func (ss *SocketServer) sendAcknowledgement(conn *net.Conn, msg *Message) {

	// if acknowledgement is enabled
	if ss.SendAcknowledgement {

		data := ss.encoder.Encode(*msg)
		if data == nil {
			return
		}

		err := wsutil.WriteServerMessage(*conn, ws.OpText, data)
		if err != nil {
			// log the error which triggered closing the connection
			AppLogger.Errorln("[sendAcknowledgement] error occurred while sending server reply: ", err)
		}
	}
}

func (ss *SocketServer) BroadcastAllGroups(msg *Message) {
	for groupId, _ := range ss.groups {
		ss.groups[groupId].createBroadcast(msg)
	}
}

func (ss *SocketServer) BroadcastToGroup(groupId string, msg *Message) {
	if val, ok := ss.groups[groupId]; ok {
		val.createBroadcast(msg)
	}
}
