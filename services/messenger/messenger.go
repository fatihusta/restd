package messenger

import (
	
	zmq "github.com/pebbe/zmq4"
	"github.com/jsommerville-untangle/golang-shared/services/logger"
)

// Channel to signal these routines to stop
var serviceShutdown = make (chan bool, 1)

func Startup() {
	logger.Info("Starting zmq messenger...\n")
	socket, err := setupZmqSocket()
	if err != nil {
		logger.Warn("Unable to setup ZMQ sockets")
	}

	logger.Info("Setting up client socket on zmq socket...\n")
	go socketClient(socket)
}

func Shutdown() {
	serviceShutdown <- true
}

// THIS IS A GOROUTINE
func socketClient(soc *zmq.Socket) {
	defer soc.Close()


	return
}

func setupZmqSocket() (soc *zmq.Socket, err error) {
	client, err := zmq.NewSocket(zmq.REQ)

	if err != nil {
		logger.Err("Unable to open ZMQ socket... %s\n", err)
		return nil, err
	}

	// TODO we should read a file created by packetd that contains a randomized
	// ZMQ port to lsiten on 
	client.Connect("tcp://localhost:5555")

	return client, nil
}