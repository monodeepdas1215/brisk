package brisk

const (

	AUTH_TYPE_NO_AUTH 	= "none"
	AUTH_TYPE_JWT	 	= "jwt_auth"
	AUTH_TYPE_SIMPLE 	= "simple_auth"

	SERVER_ACK_MSG 		= "ok"
)

type AuthType string

type Configuration struct {

	// host address to start the websocket server
	HostAddr			string

	// authentication type to choose from when clients connect
	AuthenticationType AuthType

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
	ack bool, broadcastMessagesLimit int64, maxConcurrency int,
	logLevel int, reportCaller bool) *Configuration {

	return &Configuration{
		HostAddr:            address,
		AuthenticationType:  auth,
		SendAcknowledgement: ack,
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
		BroadcastMessagesLimit:   1000,
		MaxThreadPoolConcurrency: 50000,
		LogLevel:                 ErrorLevel,
		LoggerReportCaller:       false,
	}
}