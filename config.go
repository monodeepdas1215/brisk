package brisk

const (

	AUTH_TYPE_NO_AUTH 	= "none"
	AUTH_TYPE_AUTH	 	= "auth"

	ENCODING_TYPE_JSON = "json"
	ENCODING_TYPE_MSG_PACK = "msg_pack"

)

type AuthType string

type Encoding string

type Configuration struct {

	// host address to start the websocket server
	HostAddr			string

	// authentication type to choose from when commonClients connect
	AuthenticationType AuthType

	// configure the message format to use with the server
	AcceptMessageEncoding Encoding

	// configure this to receive server acknowledgment
	SendAcknowledgement	bool

	// configure this for the max length of the broadcast channel
	BroadcastMessagesLimit int64

	// maximum number of parallel threads to execute simultaneously
	MaxThreadPoolConcurrency	int

	// maximum number of work requests to be held onto the worker pool
	RequestsBufferSize		int

	// log level for the entire application
	LogLevel				int

	// report the caller as well for the logger
	LoggerReportCaller		bool
}

// create a Configuration object with the given details
func NewServerConfiguration(address string, auth AuthType,
	ack bool, encoding Encoding, broadcastMessagesLimit int64, maxConcurrency int,
	logLevel int, reportCaller bool) *Configuration {

	return &Configuration{
		HostAddr:            address,
		AuthenticationType:  auth,
		SendAcknowledgement: ack,
		AcceptMessageEncoding: encoding,
		BroadcastMessagesLimit: broadcastMessagesLimit,
		MaxThreadPoolConcurrency: maxConcurrency,
		LogLevel: logLevel,
		LoggerReportCaller: reportCaller,
	}
}

// create a default Configuration object
func DefaultServerConfiguration() *Configuration {
	return &Configuration{
		HostAddr:                 "0.0.0.0:8080",
		AuthenticationType:       AUTH_TYPE_NO_AUTH,
		SendAcknowledgement:      false,
		AcceptMessageEncoding: 	  ENCODING_TYPE_JSON,
		BroadcastMessagesLimit:   1000,
		MaxThreadPoolConcurrency: 50000,
		LogLevel:                 ErrorLevel,
		LoggerReportCaller:       false,
	}
}