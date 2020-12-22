package messenger

import (
	"strconv"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/untangle/golang-shared/services/logger"
)

const (
	REQUEST_TIMEOUT = 2500 * time.Millisecond
	REQUEST_RETRIES = 3
)

// Channel to signal these routines to stop
var serviceShutdown = make (chan bool, 1)

func Startup() {
	logger.Info("Starting zmq messenger...\n")
	socket, socErr, poller := setupZmqSocket()
	if socErr != nil {
		logger.Warn("Unable to setup ZMQ sockets")
	}

	logger.Info("Setting up client socket on zmq socket...\n")
	go socketClient(socket, poller)
}

func Shutdown() {
	serviceShutdown <- true
}

// THIS IS A GOROUTINE
func socketClient(soc *zmq.Socket, poller *zmq.Poller) {
	defer soc.Close()

	sequence := 0
	retries_left := REQUEST_RETRIES
	var socErr error
	for retries_left > 0 {
		sequence++
		// send message
		logger.Info("Sending ", sequence, "\n")
		soc.SendMessage(sequence)

		for expect_reply := true; expect_reply; {
			// Poll socket for a reply, with timeout
			sockets, err := poller.Poll(REQUEST_TIMEOUT)
			if err != nil {
				break // Interrupted
			}

			//  Here we process a server reply and exit our loop if the
			//  reply is valid. If we didn't a reply we close the client
			//  socket and resend the request. We try a number of times
			//  before finally abandoning:

			if len(sockets) > 0 {
				//  We got a reply from the server, must match sequence
				reply, err := soc.RecvMessage(0)
				if err != nil {
					break //  Interrupted
				}
				seq, _ := strconv.Atoi(reply[0])
				if seq == sequence {
					logger.Info("Server replied OK (%s)\n", reply[0], "\n")
					retries_left = REQUEST_RETRIES
					expect_reply = false
				} else {
					logger.Err("Malformed reply from server: %s\n", reply, "\n")
				}
			} else {
				retries_left--
				if retries_left == 0 {
					logger.Err("Server seems to be offline, abandoning\n")
					break
				} else {
					logger.Warn("No response from server, retrying...\n")
					//  Old socket is confused; close it and open a new one
					soc.Close()
					soc, socErr, poller = setupZmqSocket()
					if socErr != nil {
						logger.Err("Unable to setup retry ZMQ sockets\n")
						break
					}
					//  Send request again, on new socket
					soc.SendMessage(sequence)
				}
			}

		}
	}

	return
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