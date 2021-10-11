package messenger

import (
	"errors"
	"sync"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/untangle/golang-shared/services/logger"
	prep "github.com/untangle/golang-shared/structs/protocolbuffers/PacketdReply"
	zreq "github.com/untangle/golang-shared/structs/protocolbuffers/ZMQRequest"
	"google.golang.org/protobuf/proto"
)

const (
	// RequestTimeout - How long to timeout on waiting for a reply request when polling
	RequestTimeout = 2500 * time.Millisecond
	// RequestRetries - Number of retries to try on a request before abandoning
	RequestRetries = 3
	// ClientTick - when keeping client open, how often to do a tick in the loop
	ClientTick = 1 * time.Minute

	// Packetd is ZMQRequest PACKETD service type, for sending requests to Packetd
	Packetd = zreq.ZMQRequest_PACKETD
	// Reportd is ZMQRequest REPORTD service type, for sending requests to reportd
	Reportd = zreq.ZMQRequest_REPORTD

	// TestInfo ZMQRequest TEST_INFO function type - for retrieving test info from packetd
	TestInfo = zreq.ZMQRequest_TEST_INFO
	// GetSessions - ZMQRequest GET_SESSIONS function type - for retrieving conntracks/sessions map from packetd
	GetSessions = zreq.ZMQRequest_GET_SESSIONS
)

// Channel to signal these routines to stop, waitgroup, socket and poller, socket error, and socket mutex
var serviceShutdown = make (chan struct{})
var wg sync.WaitGroup
var socket *zmq.Socket
var poller *zmq.Poller
var socErr error

// Use a mutex to lock/unlock use of the socket. In the lazy pirate reliability pattern, the socket is closed
// and recreated to attempt to connect to the server again. Since the socket is being closed/recreated, we don't want
// multiple requests trying to handle the same socket. 
var socketMutex sync.RWMutex

// Startup starts up the zmq messenger for restd
func Startup() {
	// Set up socket
	logger.Info("Starting zmq messenger...\n")
	socket, poller, socErr = setupZmqSocket()
	if socErr != nil {
		logger.Warn("Unable to setup ZMQ sockets")
	}

	// Adds socket to waitgroup and keeps client open
	logger.Info("Setting up client socket on zmq socket...\n")
	wg.Add(1)
	go keepClientOpen(&wg)
}

// Shutdown shuts down the zmq socket and waits for it to gracefully close
func Shutdown() {
	close(serviceShutdown)
	wg.Wait()
}

// keepClientOpen keeps the client open so the socket remains initialized
func keepClientOpen(waitgroup *sync.WaitGroup) {
	// Close socket and signal waitgroup it is done at function end
	defer socket.Close()
	defer waitgroup.Done()

	// Infinite for loop that ends when shutodwn is initialized
	tick := time.Tick(ClientTick)
	for {
		select {
		case <-serviceShutdown:
			logger.Info("Stop keeping client open\n")
			return
		case <-tick:
			logger.Debug("Restd client still open\n")
		}
	}
}

// SendRequestAndGetReply receives a ZMQrequest from the gin server, sends it, and sends the reply back to the gin server
func SendRequestAndGetReply(service zreq.ZMQRequest_Service, function zreq.ZMQRequest_Function) (socketReply [][]byte, err error) {
	// TODO - need mutexes?
	retriesLeft := RequestRetries
	var reply [][]byte
	var replyErr error
	// create request 
	zmqRequest := &zreq.ZMQRequest{Service: service, Function: function}
	// send message
	// TODO check socket is good
	logger.Debug("Sending ", zmqRequest, "\n")
	request, encodeErr := proto.Marshal(zmqRequest)
	if encodeErr != nil {
		return nil, errors.New("Failed to encode: " +  encodeErr.Error())
	}
	socketMutex.Lock()
	socket.SendMessage(request)
	socketMutex.Unlock()

	// Continue looping while expect_reply is still true
	for expectReply := true; expectReply; {
		// Poll socket for a reply, with timeout
		socketMutex.Lock()
		sockets, pollErr := poller.Poll(RequestTimeout)
		socketMutex.Unlock()
		if pollErr != nil {
			return nil, errors.New("Failed to poll socket: " + pollErr.Error())
		}

		//  Here we process a server reply and exit our loop if the
		//  reply is valid. If we didn't a reply we close the client
		//  socket and resend the request. We try a number of times
		//  before finally abandoning:

		if len(sockets) > 0 {
			//  We got a reply from the server, retrieve it and return on any errors
			socketMutex.Lock()
			reply, replyErr = socket.RecvMessageBytes(0)
			socketMutex.Unlock()
			if replyErr != nil {
				return nil, errors.New("Failed to receive a message: " + replyErr.Error())
			}
			// If the serverError was not packaged into a reply properly, it will be an empty byte array
			if len(reply[0]) == 0 {
				return nil, errors.New("Failed to create server error message, but there was a server error")
			}
			logger.Info("Server replied OK (%s)\n", reply[0], "\n")
			expectReply = false
		} else {
			// continue retrying until retries_left is exhausted
			retriesLeft--
			if retriesLeft == 0 {
				logger.Warn("Server seems to be offline, abandoning\n")
				return nil, errors.New("Server seems to be offline, abandoning")
			} 
			// recreate socket and try to resend
			socketMutex.Lock()
			logger.Warn("No response from server, retrying...\n")
			//  Old socket is confused; close it and open a new one
			socket.Close()
			socket, poller, socErr = setupZmqSocket()
			if socErr != nil {
				socketMutex.Unlock()
				return nil, errors.New("Unable to setup retry ZMQ sockets")
			}
			//  Send request again, on new socket
			socket.SendMessage(request)
			socketMutex.Unlock()
		}

	}

	return reply, nil
}

// setupZmqSocket sets up the zmq socket for initializaion and other retry sockets
func setupZmqSocket() (soc *zmq.Socket, clientPoller *zmq.Poller, SocErr error) {
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		logger.Err("Unable to open ZMQ socket... %s\n", err)
		return nil, nil, err
	}

	// TODO we should read a file created by packetd that contains a randomized
	// ZMQ port to lsiten on 
	client.Connect("tcp://localhost:5555")

	// Create poller for polling for results. If nothing is polled, retries are attempted
	poller := zmq.NewPoller()
	poller.Add(client, zmq.POLLIN)

	return client, poller, nil
}

// RetrievePacketdReplyItem retrieves the proper items needed from a PacketdReply
func RetrievePacketdReplyItem(msg [][]byte, function zreq.ZMQRequest_Function) ([]map[string]interface{}, error) {
	// Unencode the reply
	unencodedReply := &prep.PacketdReply{}
	unmarshalErr := proto.Unmarshal(msg[0], unencodedReply)
	if unmarshalErr != nil {
		return nil, errors.New("Failed to unencode: " + unmarshalErr.Error())
	}

	// If a serverError exists, return it
	if len(unencodedReply.ServerError) != 0 {
		return nil, errors.New(unencodedReply.ServerError)
	}

	// Based on function, set the result to the right protobuf data structure
	var result []map[string]interface{}
	resultItem := make(map[string]interface{})
	switch function {
	case GetSessions:
		resultItem["result"] = unencodedReply.Conntracks
	case TestInfo:
		resultItem["result"] = unencodedReply.TestInfo
	default:
		resultItem["result"] = nil
	}
	result = append(result, resultItem)

	return result, nil
}