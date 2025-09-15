.PHONY: build test run-server run-client fmt lint

build:
	go build ./...

test:
	go test ./...

run-server:
	cd cmd/server && go run .

run-server-lan:
	cd cmd/server && LAN=1 go run .

run-client:
	cd cmd/client && go run .

fmt:
	go fmt ./...
