#!/bin/bash
##
## Build mfw admin UI
##
TARGET=$1

docker-compose -f build/docker-compose.build.yml up --build musl-local
ssh root@$TARGET "/etc/init.d/restd stop"; 
sleep 5
scp ./cmd/restd/restd root@$TARGET:/usr/bin/; 
ssh root@$TARGET "/etc/init.d/restd start"

