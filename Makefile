
run_client:
	cd cmd/client && go run .

client:
	mkdir -pv build && cd cmd/client && go build . && mv client ../../build && cd ../..

run-server:
	cd cmd/server && go run .

server:
	mkdir -pv build && cd cmd/server && go build . && mv server ../../build && cd ../..

all: client server

