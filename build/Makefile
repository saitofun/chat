
root=$(realpath ..)
cmd=$(realpath ..)/cmd
client_cmd = $(cmd)/client
server_cmd = $(cmd)/server
build_root = $(root)/build

build: download clean client server

client: download
	cd $(client_cmd) && go build -v . && mv client $(build_root)

server: download
	cd $(server_cmd) && go build -v . && mv server $(build_root)

clean:
	rm -rf $(build_root)/client $(build_root)/server

download:
	cd $(root)
	go mod download
