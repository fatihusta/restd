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
	REQUEST_TIMEOUT = 2500 * time.Millisecond
	REQUEST_RETRIES = 3
)

// Channel to signal these routines to stop
var serviceShutdown = make (chan struct{})
var wg sync.WaitGroup
var socket *zmq.Socket
var poller *zmq.Poller
var socErr error

func Startup() {
	logger.Info("Starting zmq messenger...\n")
	socket, socErr, poller = setupZmqSocket()
	if socErr != nil {
		logger.Warn("Unable to setup ZMQ sockets")
	}

	logger.Info("Setting up client socket on zmq socket...\n")
	wg.Add(1)
	go keepClientOpen(&wg)
}

func Shutdown() {
	close(serviceShutdown)
	wg.Wait()
}

func keepClientOpen(waitgroup *sync.WaitGroup) {
	defer socket.Close()
	defer waitgroup.Done()

	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-serviceShutdown:
			logger.Info("Stop keeping client open\n")
			return
		case <-tick:
			logger.Info("Restd client still open\n")
		}
	}
}

func CreateRequest(service string, function string) *zreq.ZMQRequest {
	request := &zreq.ZMQRequest{Service: service, Function: function}

	return request
}

func SendRequestAndGetReply(requestRaw *zreq.ZMQRequest) (socketReply [][]byte, err error) {
	retries_left := REQUEST_RETRIES
	var reply [][]byte
	var replyErr error
	// send message
	logger.Info("Sending ", requestRaw, "\n")
	// TODO check socket is good
	request, encodeErr := proto.Marshal(requestRaw)
	if encodeErr != nil {
		return nil, errors.New("Failed to encode: " +  encodeErr.Error())
	}
	socket.SendMessage(request)

	for expect_reply := true; expect_reply; {
		// Poll socket for a reply, with timeout
		sockets, pollErr := poller.Poll(REQUEST_TIMEOUT)
		if pollErr != nil {
			return nil, errors.New("Failed to poll socket: " + pollErr.Error())
		}

		//  Here we process a server reply and exit our loop if the
		//  reply is valid. If we didn't a reply we close the client
		//  socket and resend the request. We try a number of times
		//  before finally abandoning:

		if len(sockets) > 0 {
			//  We got a reply from the server, must match sequence
			reply, replyErr = socket.RecvMessageBytes(0)
			if replyErr != nil {
				return nil, errors.New("Failed to receive a message: " + replyErr.Error())
			}
			logger.Info("Server replied OK (%s)\n", reply[0], "\n")
			expect_reply = false
		} else {
			retries_left--
			if retries_left == 0 {
				return nil, errors.New("Server seems to be offline, abandoning")
			} else {
				logger.Warn("No response from server, retrying...\n")
				//  Old socket is confused; close it and open a new one
				socket.Close()
				socket, socErr, poller = setupZmqSocket()
				if socErr != nil {
					return nil, errors.New("Unable to setup retry ZMQ sockets\n")
				}
				//  Send request again, on new socket
				socket.SendMessage(request)
			}
		}

	}

	return reply, nil
}

func setupZmqSocket() (soc *zmq.Socket, SocErr error, clientPoller *zmq.Poller) {
	client, err := zmq.NewSocket(zmq.REQ)

	if err != nil {
		logger.Err("Unable to open ZMQ socket... %s\n", err)
		return nil, err, nil
	}

	// TODO we should read a file created by packetd that contains a randomized
	// ZMQ port to lsiten on 
	client.Connect("tcp://localhost:5555")

	poller := zmq.NewPoller()
	poller.Add(client, zmq.POLLIN)

	return client, nil, poller
}

func RetrievePacketdReplyItem(msg [][]byte) ([]map[string]interface{}, error) {
	unencodedReply := &prep.PacketdReply{}
	if unmarshalErr := proto.Unmarshal(msg[0], unencodedReply); unmarshalErr != nil {
		return nil, errors.New("Failed to unencode: " + unmarshalErr.Error())
	}

	//var result []map[string]interface{}
	//result = append(result, unencodedReply.Conntracks)
	var result []map[string]interface{}
	resultItem := make(map[string]interface{})
	resultItem["result"] = unencodedReply.Conntracks
	result = append(result, resultItem)

	return result, nil
}