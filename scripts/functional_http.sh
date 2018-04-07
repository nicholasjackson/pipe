#!/bin/bash
docker create -v /go/src/github.com/nicholasjackson --name configs alpine:3.4 /bin/true > /dev/null
docker cp ${GOPATH}/src/github.com/nicholasjackson/pipe configs:/go/src/github.com/nicholasjackson

docker run -it \
  --volumes-from configs \
	-w /go/src/github.com/nicholasjackson/pipe/test_functional/http \
	golang \
  go test -v ./main_test.go

EXIT_CODE=$?

docker rm configs > /dev/null
