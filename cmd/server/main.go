package main

import "github.com/monodeepdas1215/brisk/pkg/server"

func main() {

	configuration := server.DefaultServerConfiguration("localhost:80")
	configuration.SetRedisHostAddr("localhost:6379")

	socketServer := server.DefaultSocketServer(configuration)
	socketServer.StartListening()

}
