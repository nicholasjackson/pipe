VERSION=0.4.11
NAMESPACE=quay.io/nicholasjackson

mocks:
	go generate ./...

test:
	GOMAXPROCS=7 go test -v -parallel 7 -cover -race `go list ./... | grep -v test_functional`

coverage:
	go test -coverprofile c.out `go list ./... | grep -v test_functional`

build:
	go build -o pipe-server .

build_linux:
	CGO_ENABLED=0 GOOS=linux go build -o pipe-server .

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

goconvey:
	goconvey -excludedDirs "test_functional,vendor"

test_nats_provider:
	./scripts/functional_nats.sh

test_http_provider:
	./scripts/functional_http.sh

test_functional: test_nats_provider test_http_provider
