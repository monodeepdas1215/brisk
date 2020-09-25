package main

const (

	ENCODING_TYPE_JSON = "json"
	ENCODING_TYPE_MSG_PACK = "msg_pack"

)

type AuthType string

type Encoding string

type Configuration struct {

	// host address to start the websocket server
	HostAddr			string

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

// create a default Configuration object
func DefaultServerConfiguration(bindAddr string) *Configuration {
	return &Configuration{
		HostAddr:                 bindAddr,
		SendAcknowledgement:      false,
		AcceptMessageEncoding:    ENCODING_TYPE_JSON,
		BroadcastMessagesLimit:   100,
		MaxThreadPoolConcurrency: 50000,
		LogLevel:                 ErrorLevel,
		LoggerReportCaller:       false,
	}
}

func (cfg *Configuration) SetSendAcknowledgment(flag bool) *Configuration {
	cfg.SendAcknowledgement = flag
	return cfg
}

func (cfg *Configuration) SetAcceptMessageEncoding(encoding Encoding) *Configuration {
	cfg.AcceptMessageEncoding = encoding
	return cfg
}

func (cfg *Configuration) SetLogLevel(logLevel int) *Configuration {
	cfg.LogLevel = logLevel
	return cfg
}

func (cfg *Configuration) SetMaxThreadPoolConcurrency(concurrency int) *Configuration {
	cfg.MaxThreadPoolConcurrency = concurrency
	return cfg
}

func (cfg *Configuration) SetBroadcastMessagesLimit(limit int64) *Configuration {
	cfg.BroadcastMessagesLimit = limit
	return cfg
}