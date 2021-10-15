#!/bin/bash
##
## Build mfw restd
##
TARGET=$1
PORT=22
LOCAL_MUSL_BUILD=false
BACKUP=false

while getopts "t:p:m:b:" flag; do
    case "${flag}" in
        t) TARGET=${OPTARG} ;;
        p) PORT=${OPTARG} ;;
        m) LOCAL_MUSL_BUILD=${OPTARG} ;;
        b) BACKUP=${OPTARG}
    esac
done
shift $((OPTIND-1))

echo "Sending package to $TARGET with port: $PORT and local musl build: $LOCAL_MUSL_BUILD and backup: $BACKUP"

if [ "$LOCAL_MUSL_BUILD" = true ]
then
    docker-compose -f build/docker-compose.build.yml up --build musl-local
else
    docker-compose -f build/docker-compose.build.yml up --build musl
fi

ssh-copy-id root@$TARGET

if [ "$BACKUP" = true ]
then
    now=`date +"%N"`
    mkdir "restd_backup"
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P $PORT root@$TARGET:/usr/bin/restd ./restd_backup/restd_${now}; 
fi


ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p $PORT root@$TARGET "/etc/init.d/restd stop"; 
sleep 5
scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P $PORT ./cmd/restd/restd root@$TARGET:/usr/bin/; 
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p $PORT root@$TARGET "/etc/init.d/restd start"
