package api

import (
	"fmt"

	"github.com/saitofun/chat/cmd/config"
	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/chat/pkg/modules/users"
	"github.com/saitofun/qlib/net/qsock"
)

var (
	server *qsock.Server
)

func init() {
	var err error
	server, err = qsock.NewServer(
		qsock.ServerOptionConnCap(10),
		qsock.ServerOptionListenAddr(config.ServerAddr),
		qsock.ServerOptionProtocol(qsock.ProtocolTCP),
		qsock.ServerOptionParser(protoc.Parser),
		// qsock.ServerOptionHandler(func(ev *qsock.Event) {
		// 	var (
		// 		err error
		// 		seq string
		// 	)
		// 	switch pl := ev.Payload().(type) {
		// 	case *protoc.Echo:
		// 		err, seq = OnEcho(pl, ev.Endpoint()), pl.ID().String()
		// 	case *protoc.Instruct:
		// 		err, seq = OnGmCmd(pl, ev.Endpoint()), pl.ID().String()
		// 	default:
		// 		err = errors.ErrUnknownGmCmd
		// 	}
		// 	fmt.Println(seq, err)
		// }),
		qsock.ServerOptionRoute(protoc.CmdEcho, OnEcho),
		qsock.ServerOptionRoute(protoc.CmdInstruct, OnGmCmd),
		qsock.ServerOptionOnDisconnected(func(n *qsock.Node) {
			users.Controller().UserOffline(n.ID())
		}),
	)
	if err != nil {
		panic(err)
	}
	go server.Serve()
	fmt.Println("Chat server started: ", config.ServerAddr)
}
