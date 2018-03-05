#!/bin/bash

# Start containers and copy local files, we need to run the tests in a container which 
# is attached to the same network as nats because of the way that circle remote docker works
docker run -d --name nats -p 4222:4222 nats-streaming:0.7.0-linux > /dev/null
sleep 10
docker create -v /go/src/github.com/nicholasjackson --name configs alpine:3.4 /bin/true > /dev/null
docker cp ${GOPATH}/src/github.com/nicholasjackson/pipe configs:/go/src/github.com/nicholasjackson

docker run -it \
	--network container:nats \
  --volumes-from configs \
	-e "nats_server=localhost" \
	-e "nats_cluster_id=test-cluster" \
	-w /go/src/github.com/nicholasjackson/pipe/test_functional/nats \
	golang \
	go test -v ./main_test.go
EXIT_CODE=$?

docker stop nats > /dev/null
docker rm nats > /dev/null
docker rm configs > /dev/null

exit $EXIT_CODE
