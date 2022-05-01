package main

import (
	"fmt"
	"os"

	_ "github.com/saitofun/chat/cmd/server/api"
)

// wait wait exit signal
func wait() {
	c := make(chan os.Signal, 1)
	_ = <-c
	fmt.Println("server exited")
	// @todo server.Shutdown() not implemented
}

func main() {
	wait()
}
