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
}

// starts with a basic configuration setting
func BasicSocketServer(config *Configuration) *SocketServer {

	if config == nil {
		config = DefaultServerConfiguration()
	}

	InitLogger(config.LogLevel, config.LoggerReportCaller)

	obj := &SocketServer{
		Configuration:   config,
		clientsHub:      newClientsHub(config.BroadcastMessagesLimit),
		ServerCallbacks: nil,
		Splash:          core.NewSplashPool(config.RequestsBufferSize, config.MaxThreadPoolConcurrency, core.ErrorLevel),
	}
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

	if ss.AuthenticationType != AUTH_TYPE_NO_AUTH && ss.AuthHandler == nil {
		panic("need to specify a AuthHandler function as the AuthenticationType is not none")
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
		ss.addClient(id, &conn)

		// trigger the callback function
		// TODO convert this to worker pool
		if ss.OnClientConnected != nil {
			go ss.OnClientConnected(id)
		}

		// handling a particular client in a new loop
		ss.Splash.AddWorkRequest(&handleMessagesWorkReq{
			clientId: id,
			conn:     &conn,
			ss: ss,
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

type handleMessagesWorkReq struct {
	clientId string
	conn *net.Conn
	ss	 *SocketServer
}

func (hm *handleMessagesWorkReq) Execute() {

	defer hm.ss.removeClient(hm.clientId, hm.ss.OnClientDisconnected)

	// waiting for the messages from client
	for {

		message, op, err := wsutil.ReadClientData(*hm.conn)
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

		AppLogger.logger.Infoln(string(message))


	}
}

func (hm *handleMessagesWorkReq) GetId() string {
	return hm.clientId + " handleMessagesWorkRequest"
}

//// handleMessagesWorkReq is the function which deals with receiving server messages from respective clients
//// and handle their authentication accordingly
//// if Auth is enabled from the server then authenticate the clients first before letting them exchange data
//// drop their connection otherwise.
//func (ss *SocketServer) handleMessagesWorkReq(clientId string, conn *net.Conn) {
//
//	defer ss.removeClient(cli)
//
//}