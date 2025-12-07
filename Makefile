.PHONY: client server run test

SERVER_DIR := src/server
CLIENT_DIR := src/client

test:
	cd $(SERVER_DIR) && go test ./...

server: test
	cd $(SERVER_DIR) && go run main.go

client:
	cd $(CLIENT_DIR) && python -m http.server 8080

run: 
	cd $(SERVER_DIR) && go test ./... &&  go run main.go &
	cd $(CLIENT_DIR) && python -m http.server 8080
