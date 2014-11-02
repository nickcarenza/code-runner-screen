VERSION := 0.0.1a
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

deps:
	go get -u -t -v ./...

godep:
	go get github.com/tools/godep

deps-save: 
	godep save -r ./...

build: 
	go build -o bin/runner -ldflags "-X main.version $(VERSION)dev-$(SHA)" runner.go