VERSION=0.4.11
NAMESPACE=quay.io/nicholasjackson

mocks:
	go generate ./...

test:
	GOMAXPROCS=7 go test -parallel 7 -cover -race ./...

build:
	go build -o pipe .

build_linux:
	CGO_ENABLED=0 GOOS=linux go build -o pipe .

build_docker: build_linux
	docker build -t ${NAMESPACE}/pipe:${VERSION} .
	docker tag ${NAMESPACE}/pipe:${VERSION} ${NAMESPACE}/pipe:${VERSION}
	docker tag ${NAMESPACE}/pipe:${VERSION} ${NAMESPACE}/pipe:latest

push_docker: 
	docker push ${NAMESPACE}/pipe:${VERSION}
	docker push ${NAMESPACE}/pipe:latest

run_docker:
	docker run -it -v $(shell pwd)/example_config.yml:/etc/pipe/example_config.yml ${NAMESPACE}/pipe:latest -config /etc/pipe/example_config.yml

build_all:
	goreleaser -snapshot -rm-dist -skip-validate

test_nats_provider:
	go test -v test_functional/nats/main_test.go
