package main

import (
	"fmt"
	"os"

	"github.com/saitofun/chat/cmd/client/api"
)

func main() {
	wait()
}

// wait wait exit signal
func wait() {
	c := make(chan os.Signal, 1)
	_ = <-c
	fmt.Println("client exited")
	if api.User != nil && api.LoginAt != nil {
		fmt.Printf("\tusername: %s\n", api.User.Name)
		fmt.Printf("\tlogin at: %s\n", api.LoginAt.String())
	}
	api.Logoff("user terminated")
	api.Shutdown()
}
