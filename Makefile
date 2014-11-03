VERSION := 0.0.1a

deps:
	go get -u -t -v ./...

godep:
	go get github.com/tools/godep

deps-save: 
	godep save -r ./...

build: 
	go build -o bin/runner -ldflags "-X main.version $(VERSION)" runner.go

install: build
	install bin/runner /usr/local/bin