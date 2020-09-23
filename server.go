package brisk

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
	"github.com/monodeepdas1215/splash/core"
	"net"
	"syscall"
)

type SocketServer struct {
	*Configuration
	*clientsHub
	*ServerCallbacks
	*core.Splash
	encoder IEncoder
}

// starts with a basic configuration setting
func BasicSocketServer(config *Configuration) *SocketServer {

	InitLogger(config.LogLevel, config.LoggerReportCaller)

	obj := &SocketServer{
		Configuration:   config,
		clientsHub:      newClientsHub(config.BroadcastMessagesLimit),
		ServerCallbacks: nil,
		Splash:          core.NewSplashPool(config.RequestsBufferSize, config.MaxThreadPoolConcurrency, core.ErrorLevel),
	}

	var appEncoder IEncoder

	switch obj.AcceptMessageEncoding {
	case ENCODING_TYPE_JSON:
		appEncoder = &JsonEncoder{}
	case ENCODING_TYPE_MSG_PACK:
		panic("encoding using msg_pack has not been implemented yet")
	}
	obj.encoder = appEncoder

	return obj
}

// sets the callback functions
func (ss *SocketServer) SetCallbackFunctions(callbacks *ServerCallbacks) *SocketServer {
	ss.ServerCallbacks = callbacks
	return ss
}

// does some initial checks to ensure the presence of all handler functions in place.
// increases the ulimit and starts the server loop
func (ss *SocketServer) StartListening() {

	// doing all the checks here before starting
	if ss.ServerCallbacks == nil {
		panic("server cannot start without the callbacks initialized to valid handler functions")
	}

	// increasing ulimit to accept more than 1024 connections
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
			AppLogger.logger.Errorln("error occurred while listening to next connection: ", err)
			return
		}

		AppLogger.logger.Infoln("incoming connection from: ", conn.RemoteAddr())

		// upgrade the tcp connection to websocket protocol
		handshakeObj, err := ss.tcpConnectionUpgrader().Upgrade(conn)
		if err != nil {
			AppLogger.logger.Errorln("error occurred while upgrading TCP connection to websocket: ", err)
			continue
		}

		AppLogger.logger.Infof("connection from Client(%s) upgraded successfully. HandshakeObj: ", conn.RemoteAddr(), handshakeObj)

		id := uuid.New().String()

		incomingClient := newSocketClient(id, &conn)
		ss.addClient(incomingClient)
		
		if ss.OnClientConnected != nil {
			ss.OnClientConnected(id)
		}

		// handling a particular incomingClient in a new loop
		ss.Splash.AddWorkRequest(&handleMessagesWorkReq{
			client: incomingClient,
			ss:     ss,
		})
	}
}

// setups the tcp connection for the given address
func (ss *SocketServer) setupTCPConnection() (net.Listener, error) {

	listener, err := net.Listen("tcp", ss.HostAddr)
	if err != nil {
		AppLogger.logger.Errorln("error occurred while listening to TCP connections: ", err)
		return nil, err
	}

	AppLogger.logger.Infoln("server listening for TCP connections on the address: ", ss.HostAddr)
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
			AppLogger.logger.Errorln("error occurred while sending server reply: ", err)
		}
	}
}

func (ss *SocketServer) SendMessage(asyncSend bool, clientId string, msg Message, opCode ws.OpCode) {

	if asyncSend {

		ss.AddWorkRequest(&sendMessageWorkReq{
			clientId: clientId,
			code:     opCode,
			ss:       ss,
			msg:      msg,
		})

	} else {
		data := ss.encoder.Encode(msg)
		ss.pushToClient(clientId, data, opCode)
	}

}