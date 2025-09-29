.PHONY: build test run-server run-client fmt stop

PORT ?= 9090

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

stop:
	@PID=$$(lsof -t -nP -iTCP:$(PORT) -sTCP:LISTEN 2>/dev/null || true); \
	if [ -n "$$PID" ]; then \
		kill $$PID && echo "Killed process $$PID on port $(PORT)"; \
	else \
		echo "No process running on port $(PORT)"; \
	fi
