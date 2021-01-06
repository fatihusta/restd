# restd
RESTful API and web server daemon

Building in docker
==================

MUSL target
-----------

```
docker-compose -f build/docker-compose.build.yml up --build musl-local
```

glibc target
-----------

```
docker-compose -f build/docker-compose.build.yml up --build glibc-local
```

RESTD and ZMQ
=============

When adding a new restd endpoint to restd that requires information from packetd or reportd (the 'server'), ZMQ will be needed. 
When doing this, keep in mind the following: 

1. In golang-shared, add the function that you'll be using for the server into ZMQRequest.proto. Add the data type into the reply message i.e. PacketdReply. Make sure to build the messages before committing! 
2. In restd, add the endpoint. In the messenger service, add the function type as a constant. In the retrieve function i.e. RetrievePacketdReplyItem, add to the switch statement for retrieving the right information to send to the gin server. In the gind service, add the functions to call the messenger SendRequestAndGetReply, the retrieval function, and sending it to the front end. 
3. In the server i.e. packetd, in the zmq service, define the new function type as a constant. In the Process function, add to the switch statement the logic needed to retrieve the information from the server and package it into a zmq protobuf reply message. 
