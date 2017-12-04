VERSION=0.4.7
NAMESPACE=quay.io/nicholasjackson

mocks:
	go generate ./...

test:
	GOMAXPROCS=7 go test -parallel 7 -cover -race ./...

build:
	go build -o faas-nats .

build_linux:
	CGO_ENABLED=0 GOOS=linux go build -o faas-nats .

build_docker: build_linux
	docker build -t ${NAMESPACE}/faas-nats:${VERSION} .
	docker tag ${NAMESPACE}/faas-nats:${VERSION} ${NAMESPACE}/faas-nats:${VERSION}
	docker tag ${NAMESPACE}/faas-nats:${VERSION} ${NAMESPACE}/faas-nats:latest

push_docker: 
	docker push ${NAMESPACE}/faas-nats:${VERSION}
	docker push ${NAMESPACE}/faas-nats:latest

run_docker:
	docker run -it -v $(shell pwd)/example_config.yml:/etc/faas-nats/example_config.yml ${NAMESPACE}/faas-nats:latest -config /etc/faas-nats/example_config.yml
